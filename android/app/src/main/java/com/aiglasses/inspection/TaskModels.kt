// 眼镜端面向业务的 Kotlin 模型。这些类型对应后端任务和节点概念，包含必填证据校验和
// 幂等键，用于弱网现场环境下安全重试节点结果提交。
package com.aiglasses.inspection

import java.util.UUID

data class InspectionTask(
    val id: Long,
    val title: String,
    val dueAt: String,
    val nodes: List<InspectionNode>
)

data class InspectionNode(
    val id: Long,
    val name: String,
    val minPhotos: Int,
    val requireText: Boolean,
    val allowAbnormal: Boolean,
    val status: NodeStatus = NodeStatus.Pending
)

enum class NodeStatus { Pending, Completed, Abnormal }

data class PendingNodeResult(
    val taskId: Long,
    val nodeId: Long,
    val idempotencyKey: String = UUID.randomUUID().toString(),
    val textNote: String,
    val attachmentIds: List<Long>,
    val abnormal: Boolean
)

// canSubmit 根据节点要求判断文字说明和附件数量是否满足提交条件。
fun InspectionNode.canSubmit(textNote: String, attachmentIds: List<Long>): Boolean {
    if (requireText && textNote.isBlank()) return false
    if (attachmentIds.size < minPhotos) return false
    return true
}
