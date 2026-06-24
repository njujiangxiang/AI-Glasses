#!/usr/bin/env python3
import re
import xml.etree.ElementTree as ET
from collections import defaultdict
from dataclasses import dataclass
from typing import Dict, List, Optional, Set, Tuple


@dataclass
class PDMColumn:
    code: str
    name: str
    data_type: str
    length: Optional[int] = None
    precision: Optional[int] = None
    mandatory: bool = False
    default_value: Optional[str] = None
    comment: Optional[str] = None


@dataclass
class PDMTable:
    code: str
    name: str
    columns: Dict[str, PDMColumn]
    primary_keys: List[str]
    indexes: Dict[str, List[str]]


@dataclass
class GoField:
    name: str
    go_type: str
    gorm_tags: Dict[str, str]
    column_name: str
    is_primary_key: bool = False
    is_nullable: bool = True
    size: Optional[int] = None
    unique: bool = False
    index: bool = False
    comment: Optional[str] = None


@dataclass
class GoModel:
    struct_name: str
    table_name: str
    fields: Dict[str, GoField]


def parse_pdm(file_path: str) -> Dict[str, PDMTable]:
    """解析 PowerDesigner PDM XML 文件"""
    # 解析 XML
    tree = ET.parse(file_path)
    root = tree.getroot()

    tables = {}
    column_map = {}  # 用于通过 ID 查找列

    # 首先建立列的 ID 映射
    for col_elem in root.findall(".//{object}Column"):
        col_id = col_elem.get("Id")
        if col_id:
            code_elem = col_elem.find("./{attribute}Code")
            if code_elem is not None and code_elem.text:
                column_map[col_id] = code_elem.text.strip()

    # 查找所有 Table 元素
    for table_elem in root.findall(".//{object}Table"):
        # 提取表信息
        code_elem = table_elem.find("./{attribute}Code")
        name_elem = table_elem.find("./{attribute}Name")

        if code_elem is None:
            continue

        table_code = code_elem.text.strip() if code_elem.text else ""
        table_name = name_elem.text.strip() if (name_elem is not None and name_elem.text) else table_code

        # 提取表注释（从 PhysicalOptions）
        table_comment = ""
        phys_opts_elem = table_elem.find("./{attribute}PhysicalOptions")
        if phys_opts_elem is not None and phys_opts_elem.text:
            import re
            comment_match = re.search(r"COMMENT='([^']*)'", phys_opts_elem.text)
            if comment_match:
                table_comment = comment_match.group(1)

        columns = {}
        primary_keys = []
        indexes = {}

        # 提取列
        columns_elem = table_elem.find("./{collection}Columns")
        if columns_elem is not None:
            for col_elem in columns_elem.findall("./{object}Column"):
                col_code_elem = col_elem.find("./{attribute}Code")
                if col_code_elem is None:
                    continue

                col_code = col_code_elem.text.strip() if col_code_elem.text else ""
                col_name_elem = col_elem.find("./{attribute}Name")
                col_name = col_name_elem.text.strip() if (col_name_elem is not None and col_name_elem.text) else col_code

                data_type_elem = col_elem.find("./{attribute}DataType")
                data_type = data_type_elem.text.strip() if (data_type_elem is not None and data_type_elem.text) else ""

                length_elem = col_elem.find("./{attribute}Length")
                length = int(length_elem.text) if (length_elem is not None and length_elem.text) else None

                precision_elem = col_elem.find("./{attribute}Precision")
                precision = int(precision_elem.text) if (precision_elem is not None and precision_elem.text) else None

                # 注意：PDM 中用的是 Column.Mandatory
                mandatory_elem = col_elem.find("./{attribute}Column.Mandatory")
                mandatory = (mandatory_elem.text.strip() == "1") if (mandatory_elem is not None and mandatory_elem.text) else False

                default_elem = col_elem.find("./{attribute}DefaultValue")
                default_value = default_elem.text.strip() if (default_elem is not None and default_elem.text) else None

                comment_elem = col_elem.find("./{attribute}Comment")
                comment = comment_elem.text.strip() if (comment_elem is not None and comment_elem.text) else None

                columns[col_code.lower()] = PDMColumn(
                    code=col_code,
                    name=col_name,
                    data_type=data_type,
                    length=length,
                    precision=precision,
                    mandatory=mandatory,
                    default_value=default_value,
                    comment=comment
                )

        # 提取主键
        primary_key_elem = table_elem.find("./{collection}PrimaryKey")
        if primary_key_elem is not None:
            for key_ref in primary_key_elem.findall("./{object}Key"):
                # 需要通过 Key 找到关联的列
                key_id = key_ref.get("Ref")
                if key_id:
                    # 查找 Key 定义
                    key_elem = root.find(f".//{{object}}Key[@Id='{key_id}']")
                    if key_elem is not None:
                        key_columns_elem = key_elem.find("./{collection}Key.Columns")
                        if key_columns_elem is not None:
                            for col_ref in key_columns_elem.findall("./{object}Column"):
                                col_id = col_ref.get("Ref")
                                if col_id in column_map:
                                    primary_keys.append(column_map[col_id].lower())

        # 提取索引
        indexes_elem = table_elem.find("./{collection}Indexes")
        if indexes_elem is not None:
            for idx_elem in indexes_elem.findall("./{object}Index"):
                idx_code_elem = idx_elem.find("./{attribute}Code")
                idx_name_elem = idx_elem.find("./{attribute}Name")
                idx_code = idx_code_elem.text.strip() if (idx_code_elem is not None and idx_code_elem.text) else ""
                idx_name = idx_name_elem.text.strip() if (idx_name_elem is not None and idx_name_elem.text) else idx_code

                idx_columns = []

                idx_columns_elem = idx_elem.find("./{collection}IndexColumns")
                if idx_columns_elem is not None:
                    for idx_col_elem in idx_columns_elem.findall("./{object}IndexColumn"):
                        col_ref_elem = idx_col_elem.find("./{collection}Column")
                        if col_ref_elem is not None:
                            col_ref = col_ref_elem.find("./{object}Column")
                            if col_ref is not None:
                                col_id = col_ref.get("Ref")
                                if col_id in column_map:
                                    idx_columns.append(column_map[col_id].lower())

                if idx_columns:
                    indexes[idx_name.lower()] = idx_columns

        tables[table_code.lower()] = PDMTable(
            code=table_code,
            name=table_comment if table_comment else table_name,
            columns=columns,
            primary_keys=primary_keys,
            indexes=indexes
        )

    return tables


def parse_go_models(file_path: str) -> Dict[str, GoModel]:
    """解析 Go 模型文件"""
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    models = {}

    # 查找所有 struct 定义
    struct_pattern = re.compile(
        r'type\s+(\w+)\s+struct\s*\{([^}]*)\}',
        re.MULTILINE | re.DOTALL
    )

    for match in struct_pattern.finditer(content):
        struct_name = match.group(1)
        struct_body = match.group(2)

        # 推断表名（通常是复数形式或小写）
        # 简单规则：struct 名转小写
        table_name = struct_name.lower()

        # 解析字段
        fields = {}

        # 字段模式：FieldName Type `gorm:"..." json:"..."`
        field_pattern = re.compile(
            r'(\w+)\s+([\w\[\]\*]+)\s*`([^`]*)`',
            re.MULTILINE
        )

        for field_match in field_pattern.finditer(struct_body):
            field_name = field_match.group(1)
            go_type = field_match.group(2)
            tags_str = field_match.group(3)

            # 解析标签
            gorm_tags = {}
            column_name = snake_case(field_name)  # 默认是 snake_case

            # 解析 gorm 标签
            gorm_match = re.search(r'gorm:"([^"]*)"', tags_str)
            if gorm_match:
                gorm_parts = gorm_match.group(1).split(';')
                for part in gorm_parts:
                    part = part.strip()
                    if ':' in part:
                        key, value = part.split(':', 1)
                        gorm_tags[key.strip()] = value.strip()
                    else:
                        gorm_tags[part] = part

                # 提取 column 名
                if 'column' in gorm_tags:
                    column_name = gorm_tags['column']

            # 检查是否是主键
            is_primary_key = 'primaryKey' in gorm_tags or 'primarykey' in gorm_tags or 'primary_key' in gorm_tags

            # 检查是否非空
            is_nullable = not ('not null' in gorm_tags or 'not_null' in gorm_tags)

            # 检查大小
            size = None
            if 'size' in gorm_tags:
                try:
                    size = int(gorm_tags['size'])
                except ValueError:
                    pass

            # 检查唯一约束
            unique = 'unique' in gorm_tags or 'uniqueIndex' in gorm_tags

            # 检查索引
            index = 'index' in gorm_tags or 'uniqueIndex' in gorm_tags

            # 提取注释
            comment = gorm_tags.get('comment')

            fields[column_name.lower()] = GoField(
                name=field_name,
                go_type=go_type,
                gorm_tags=gorm_tags,
                column_name=column_name,
                is_primary_key=is_primary_key,
                is_nullable=is_nullable,
                size=size,
                unique=unique,
                index=index,
                comment=comment
            )

        # 如果没有解析到字段（可能是没有标签的简单 struct），尝试简单解析
        if not fields:
            simple_field_pattern = re.compile(
                r'(\w+)\s+([\w\[\]\*]+)',
                re.MULTILINE
            )
            for field_match in simple_field_pattern.finditer(struct_body):
                field_name = field_match.group(1)
                go_type = field_match.group(2)
                # 跳过一些特殊字段
                if field_name in ['ID', 'CreatedAt', 'UpdatedAt', 'DeletedAt']:
                    column_name = snake_case(field_name)
                    fields[column_name.lower()] = GoField(
                        name=field_name,
                        go_type=go_type,
                        gorm_tags={},
                        column_name=column_name,
                        is_primary_key=field_name == 'ID',
                        is_nullable=go_type.startswith('*') or go_type.startswith('[]'),
                        size=None,
                        unique=False,
                        index=False,
                        comment=None
                    )

        models[table_name.lower()] = GoModel(
            struct_name=struct_name,
            table_name=table_name,
            fields=fields
        )

    return models


def snake_case(s: str) -> str:
    """将驼峰命名转换为下划线命名"""
    s = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', s)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s).lower()


def map_pdm_type_to_go(pdm_type: str) -> Tuple[str, Optional[int]]:
    """将 PDM 数据类型映射到 Go 类型"""
    pdm_type_lower = pdm_type.lower()

    # 整数类型
    if 'int' in pdm_type_lower:
        if 'bigint' in pdm_type_lower or 'int64' in pdm_type_lower:
            return 'int64', None
        elif 'tinyint' in pdm_type_lower:
            return 'int8', None
        elif 'smallint' in pdm_type_lower:
            return 'int16', None
        else:
            return 'int', None

    # 字符串类型
    elif 'varchar' in pdm_type_lower or 'nvarchar' in pdm_type_lower:
        # 提取长度
        len_match = re.search(r'\((\d+)\)', pdm_type)
        length = int(len_match.group(1)) if len_match else None
        return 'string', length

    elif 'char' in pdm_type_lower:
        len_match = re.search(r'\((\d+)\)', pdm_type)
        length = int(len_match.group(1)) if len_match else None
        return 'string', length

    # 文本类型
    elif 'text' in pdm_type_lower or 'clob' in pdm_type_lower:
        return 'string', None

    # 浮点数/小数
    elif 'decimal' in pdm_type_lower or 'numeric' in pdm_type_lower:
        return 'float64', None
    elif 'float' in pdm_type_lower or 'double' in pdm_type_lower:
        return 'float64', None

    # 日期时间
    elif 'datetime' in pdm_type_lower or 'timestamp' in pdm_type_lower:
        return 'time.Time', None
    elif 'date' in pdm_type_lower:
        return 'time.Time', None

    # 布尔类型
    elif 'bit' in pdm_type_lower or 'bool' in pdm_type_lower or 'boolean' in pdm_type_lower:
        return 'bool', None

    # 二进制类型
    elif 'blob' in pdm_type_lower or 'binary' in pdm_type_lower:
        return '[]byte', None

    # 默认
    return 'string', None


def normalize_table_name(name: str) -> str:
    """标准化表名，用于匹配"""
    # 移除下划线，转小写
    return name.replace("_", "").lower()


def compare_tables(pdm_tables: Dict[str, PDMTable], go_models: Dict[str, GoModel]) -> Dict:
    """对比表"""
    # 建立标准化名称映射
    pdm_normalized = {normalize_table_name(name): (name, table) for name, table in pdm_tables.items()}
    go_normalized = {normalize_table_name(name): (name, model) for name, model in go_models.items()}

    common_normalized = set(pdm_normalized.keys()) & set(go_normalized.keys())

    tables_only_in_pdm = []
    tables_only_in_go = []
    common_tables = []

    # 找出共同表和独有表
    for norm_name in common_normalized:
        pdm_orig, pdm_table = pdm_normalized[norm_name]
        go_orig, go_model = go_normalized[norm_name]
        common_tables.append((pdm_orig, go_orig, pdm_table, go_model))

    for name in pdm_tables.keys():
        if normalize_table_name(name) not in common_normalized:
            tables_only_in_pdm.append(name)

    for name in go_models.keys():
        if normalize_table_name(name) not in common_normalized:
            tables_only_in_go.append(name)

    return {
        'tables_only_in_pdm': tables_only_in_pdm,
        'tables_only_in_go': tables_only_in_go,
        'common_tables': common_tables
    }


def compare_columns(pdm_table: PDMTable, go_model: GoModel) -> Dict:
    """对比字段"""
    pdm_column_names = set(pdm_table.columns.keys())
    go_column_names = set(go_model.fields.keys())

    columns_only_in_pdm = pdm_column_names - go_column_names
    columns_only_in_go = go_column_names - pdm_column_names
    common_columns = pdm_column_names & go_column_names

    type_differences = []
    mandatory_differences = []
    length_differences = []

    for col_name in common_columns:
        pdm_col = pdm_table.columns[col_name]
        go_field = go_model.fields[col_name]

        # 类型对比
        expected_go_type, expected_length = map_pdm_type_to_go(pdm_col.data_type)

        # 检查类型兼容性
        go_type_clean = go_field.go_type.replace('*', '').replace('[]', '')

        # 简化的类型对比
        type_compatible = False
        if go_type_clean == expected_go_type:
            type_compatible = True
        elif expected_go_type == 'int64' and go_type_clean in ['int', 'int32', 'uint', 'uint32', 'uint64']:
            type_compatible = True
        elif expected_go_type == 'int' and go_type_clean in ['int8', 'int16', 'int32', 'int64', 'uint', 'uint8', 'uint16', 'uint32', 'uint64']:
            type_compatible = True
        elif expected_go_type == 'time.Time' and go_type_clean == 'time.Time':
            type_compatible = True
        elif expected_go_type == 'string' and go_type_clean == 'string':
            type_compatible = True
        elif expected_go_type == 'float64' and go_type_clean in ['float32', 'float64']:
            type_compatible = True
        elif expected_go_type == 'bool' and go_type_clean == 'bool':
            type_compatible = True
        elif expected_go_type == '[]byte' and go_type_clean == '[]byte':
            type_compatible = True
        # 如果是指针类型，也认为兼容
        elif go_field.go_type.startswith('*'):
            underlying = go_field.go_type[1:]
            if underlying == expected_go_type:
                type_compatible = True

        if not type_compatible:
            type_differences.append({
                'column': col_name,
                'pdm_type': pdm_col.data_type,
                'go_type': go_field.go_type,
                'expected_go_type': expected_go_type
            })

        # 长度对比
        if pdm_col.length is not None and go_field.size is not None:
            if pdm_col.length != go_field.size:
                length_differences.append({
                    'column': col_name,
                    'pdm_length': pdm_col.length,
                    'go_length': go_field.size
                })

        # 非空约束对比
        if pdm_col.mandatory == go_field.is_nullable:
            # PDM mandatory=True 意味着 NOT NULL，Go 中 is_nullable=False
            if pdm_col.mandatory and go_field.is_nullable:
                mandatory_differences.append({
                    'column': col_name,
                    'pdm_mandatory': pdm_col.mandatory,
                    'go_nullable': go_field.is_nullable,
                    'issue': 'PDM 要求非空，但 Go 模型允许为空'
                })
            elif not pdm_col.mandatory and not go_field.is_nullable:
                mandatory_differences.append({
                    'column': col_name,
                    'pdm_mandatory': pdm_col.mandatory,
                    'go_nullable': go_field.is_nullable,
                    'issue': 'PDM 允许为空，但 Go 模型要求非空'
                })

    return {
        'columns_only_in_pdm': columns_only_in_pdm,
        'columns_only_in_go': columns_only_in_go,
        'type_differences': type_differences,
        'mandatory_differences': mandatory_differences,
        'length_differences': length_differences
    }


def generate_markdown_report(pdm_tables: Dict[str, PDMTable], go_models: Dict[str, GoModel], comparison: Dict) -> str:
    """生成 Markdown 格式报告"""
    lines = []

    lines.append("# PDM 与 Go 模型对比分析报告")
    lines.append("")
    lines.append("---")
    lines.append("")

    # 1. 摘要
    lines.append("## 1. 摘要")
    lines.append("")
    lines.append(f"- PDM 中表数量: {len(pdm_tables)}")
    lines.append(f"- Go 模型数量: {len(go_models)}")
    lines.append(f"- 共同表数量: {len(comparison['common_tables'])}")
    lines.append(f"- PDM 独有表数量: {len(comparison['tables_only_in_pdm'])}")
    lines.append(f"- Go 独有表数量: {len(comparison['tables_only_in_go'])}")
    lines.append("")

    # 2. 表对比
    lines.append("## 2. 表对比")
    lines.append("")

    if comparison['tables_only_in_pdm']:
        lines.append("### 2.1 PDM 中有但代码中没有的表")
        lines.append("")
        lines.append("| PDM 表名 | 注释 |")
        lines.append("|---------|------|")
        for table_name in sorted(comparison['tables_only_in_pdm']):
            table = pdm_tables[table_name]
            lines.append(f"| {table.code} | {table.name} |")
        lines.append("")

    if comparison['tables_only_in_go']:
        lines.append("### 2.2 代码中有但 PDM 中没有的表")
        lines.append("")
        lines.append("| 表名 | Struct 名 |")
        lines.append("|------|-----------|")
        for table_name in sorted(comparison['tables_only_in_go']):
            model = go_models[table_name]
            lines.append(f"| {model.table_name} | {model.struct_name} |")
        lines.append("")

    # 3. 字段对比
    lines.append("## 3. 字段对比（共同表）")
    lines.append("")

    for pdm_name, go_name, pdm_table, go_model in sorted(comparison['common_tables'], key=lambda x: x[0]):
        col_comparison = compare_columns(pdm_table, go_model)

        lines.append(f"### 3.1 表: {pdm_table.code} ↔ {go_model.struct_name}")
        lines.append(f"")
        lines.append(f"> 注释: {pdm_table.name}")
        lines.append("")

        has_issue = False

        if col_comparison['columns_only_in_pdm']:
            has_issue = True
            lines.append("#### 3.1.1 PDM 有但代码没有的字段")
            lines.append("")
            lines.append("| 字段名 | 类型 | 注释 | 非空 |")
            lines.append("|--------|------|------|------|")
            for col_name in sorted(col_comparison['columns_only_in_pdm']):
                col = pdm_table.columns[col_name]
                lines.append(f"| {col.code} | {col.data_type} | {col.comment or col.name} | {'是' if col.mandatory else '否'} |")
            lines.append("")

        if col_comparison['columns_only_in_go']:
            has_issue = True
            lines.append("#### 3.1.2 代码有但 PDM 没有的字段")
            lines.append("")
            lines.append("| 字段名 | Go 类型 | 列名 |")
            lines.append("|--------|---------|------|")
            for col_name in sorted(col_comparison['columns_only_in_go']):
                field = go_model.fields[col_name]
                lines.append(f"| {field.name} | {field.go_type} | {field.column_name} |")
            lines.append("")

        if col_comparison['type_differences']:
            has_issue = True
            lines.append("#### 3.1.3 类型差异")
            lines.append("")
            lines.append("| 字段名 | PDM 类型 | Go 类型 | 预期 Go 类型 |")
            lines.append("|--------|----------|---------|--------------|")
            for diff in col_comparison['type_differences']:
                lines.append(f"| {diff['column']} | {diff['pdm_type']} | {diff['go_type']} | {diff['expected_go_type']} |")
            lines.append("")

        if col_comparison['mandatory_differences']:
            has_issue = True
            lines.append("#### 3.1.4 非空约束差异")
            lines.append("")
            lines.append("| 字段名 | PDM 非空 | Go 可空 | 问题 |")
            lines.append("|--------|----------|---------|------|")
            for diff in col_comparison['mandatory_differences']:
                lines.append(f"| {diff['column']} | {'是' if diff['pdm_mandatory'] else '否'} | {'是' if diff['go_nullable'] else '否'} | {diff['issue']} |")
            lines.append("")

        if col_comparison['length_differences']:
            has_issue = True
            lines.append("#### 3.1.5 长度差异")
            lines.append("")
            lines.append("| 字段名 | PDM 长度 | Go 长度 |")
            lines.append("|--------|----------|---------|")
            for diff in col_comparison['length_differences']:
                lines.append(f"| {diff['column']} | {diff['pdm_length']} | {diff['go_length']} |")
            lines.append("")

        if not has_issue:
            lines.append("✅ 表结构完全一致")
            lines.append("")

    # 4. 更新建议
    lines.append("## 4. 更新建议")
    lines.append("")

    if comparison['tables_only_in_pdm']:
        lines.append("### 4.1 需要同步到代码的 PDM 表")
        lines.append("")
        for table_name in sorted(comparison['tables_only_in_pdm']):
            table = pdm_tables[table_name]
            lines.append(f"- **{table.code}** ({table.name})")
            lines.append("  ```go")
            lines.append(f"  type {table.code} struct {{")
            for col_name, col in sorted(table.columns.items()):
                go_type, _ = map_pdm_type_to_go(col.data_type)
                go_field_name = ''.join(word.capitalize() for word in col_name.split('_'))
                null_marker = ""
                if not col.mandatory and go_type not in ['[]byte']:
                    null_marker = "*"
                tags = []
                if col_name.lower() in [c.lower() for c in table.primary_keys]:
                    tags.append("primaryKey")
                if col.length:
                    tags.append(f"size:{col.length}")
                if col.mandatory:
                    tags.append("not null")
                if col.comment:
                    tags.append(f"comment:{col.comment}")
                tag_str = f' `gorm:"{";".join(tags)}"`' if tags else ''
                lines.append(f"      {go_field_name} {null_marker}{go_type}{tag_str}")
            lines.append("  }")
            lines.append("  ```")
            lines.append("")

    if comparison['tables_only_in_go']:
        lines.append("### 4.2 需要同步到 PDM 的 Go 模型表")
        lines.append("")
        for table_name in sorted(comparison['tables_only_in_go']):
            model = go_models[table_name]
            lines.append(f"- **{model.table_name}** ({model.struct_name})")
        lines.append("")

    # 5. 完整清单
    lines.append("## 5. 完整清单")
    lines.append("")

    lines.append("### 5.1 PDM 表完整清单")
    lines.append("")
    for table_name in sorted(pdm_tables.keys()):
        table = pdm_tables[table_name]
        lines.append(f"- {table.code}: {table.name} ({len(table.columns)} 列)")
    lines.append("")

    lines.append("### 5.2 Go 模型完整清单")
    lines.append("")
    for table_name in sorted(go_models.keys()):
        model = go_models[table_name]
        lines.append(f"- {model.table_name}: {model.struct_name} ({len(model.fields)} 字段)")
    lines.append("")

    return "\n".join(lines)


def main():
    pdm_path = r"D:\weChat\xwechat_files\wxid_3kbvou0dwp6w22_c269\msg\file\2026-06\AIGlasses-v0.2.pdm"
    go_path = r"d:\project\AI-Glasses\server\internal\platform\database\models.go"
    output_path = r"d:\project\AI-Glasses\pdm_go_comparison_report.md"

    print("解析 PDM 文件...")
    pdm_tables = parse_pdm(pdm_path)
    print(f"  解析到 {len(pdm_tables)} 个表")

    print("解析 Go 模型文件...")
    go_models = parse_go_models(go_path)
    print(f"  解析到 {len(go_models)} 个模型")

    print("对比分析...")
    comparison = compare_tables(pdm_tables, go_models)

    print("生成报告...")
    report = generate_markdown_report(pdm_tables, go_models, comparison)

    with open(output_path, 'w', encoding='utf-8') as f:
        f.write(report)

    print(f"报告已生成: {output_path}")


if __name__ == "__main__":
    main()
