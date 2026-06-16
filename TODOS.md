# TODOs

## Business code follow-ups

### 1. Integrate `GenerateDaily` into the first real business object

**What:** Integrate `businesscodes.Service.GenerateDaily` into the first real business object, likely task sheet or defect number.

**Why:** The first 业务编码配置 PR intentionally ships the configuration module and generator as infrastructure. A real consumer proves the service contract outside the admin “生成/试生成” path and prevents shelf-ware.

**Pros:**
- Validates the generator against real create-flow requirements.
- Forces a decision on where generated numbers are persisted and displayed.
- Catches transaction/gap expectations before more modules depend on the API.

**Cons:**
- Requires choosing the first business domain and updating that create flow.
- May expose additional product decisions, like whether task sheets or defects own the first code rule.

**Context:** After 业务编码配置 lands, choose one business object, add a code rule such as `TK`, call `GenerateDaily(ctx, "TK")` during creation, persist the resulting number, and test rollback/gap behavior. Generated Redis numbers are non-transactional; if a DB transaction rolls back after generation, gaps are acceptable unless the business explicitly rejects gaps.

**Depends on / blocked by:** Business-code configuration module merged.

### 2. Add audit logging for business-code rule changes

**What:** Add `AuditLog` entries for create, update, enable, disable, and delete operations on business-code rules.

**Why:** Business-code rules control visible identifiers. Once task sheets, defects, or devices consume `GenerateDaily`, changing code, padding, separator, or status affects future business records and should be traceable.

**Pros:**
- Gives administrators a durable trail for numbering-rule changes.
- Helps debug mismatched or surprising business numbers later.
- Reuses the existing `AuditLog` model instead of introducing a new audit system.

**Cons:**
- Adds handler/service plumbing and actor context to config operations.
- Not urgent before any real business module consumes generated numbers.

**Context:** The v1 implementation intentionally skips audit to keep the first PR focused. Revisit after the first real consumer integration, when generated numbers become business records rather than admin-only test output.

**Depends on / blocked by:** Business-code module merged; ideally first consumer integration completed.

### 3. Add an operational recovery path for sequence overflow

**What:** Add an admin/operator recovery path for sequence overflow, such as increasing padding guidance, a safe runbook, or controlled Redis key reset.

**Why:** If `seq_padding=4`, the 10000th generated number should be rejected. After that Redis keeps incrementing, so generation remains blocked until the next day unless an operator intervenes.

**Pros:**
- Avoids undocumented manual Redis operations during an outage.
- Makes overflow behavior understandable for administrators.
- Can be designed with duplicate-number safeguards once real volume is known.

**Cons:**
- Reset controls can create duplicate numbers if misused.
- Overflow may never happen if padding is chosen correctly.

**Context:** V1 should reject overflow safely with a clear error. Do not add reset controls until real usage proves the need. If urgent recovery is needed before this TODO is implemented, prefer increasing padding for future days and carefully evaluating whether deleting `BNO:{code}:{yyyyMMdd}` could duplicate already-issued numbers.

**Depends on / blocked by:** Business-code module merged and real volume observed.

### 4. Revisit the real-generation action label after first usage

**What:** Revisit the generated-number action label after first user feedback; consider renaming “试生成” to “生成下一号” or splitting dry-run preview from real generation.

**Why:** The admin action consumes a real Redis sequence. If users read “试生成” as harmless preview, they may create unexpected gaps.

**Pros:**
- Protects user trust by making sequence-consuming behavior explicit.
- Gives the team a concrete UX fix if warning copy is not enough.
- Keeps v1 small while acknowledging the naming risk.

**Cons:**
- May not be needed if the UI warning is clear.
- Could create terminology churn if changed too early.

**Context:** V1 keeps local form preview separate from real generation. The real generation action must warn that it consumes a serial number. If users still click it casually, rename the action or split dry-run and real generation more clearly.

**Depends on / blocked by:** User feedback after 业务编码配置 lands.
