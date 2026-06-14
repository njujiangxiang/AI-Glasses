// MVP 阶段 Android 眼镜端基础 Activity。它渲染简单的任务详情视图和必拍照/必填文字标记，
// 便于在接入目标眼镜硬件的相机、上传队列和网络同步之前先验证执行模型。
package com.aiglasses.inspection

import android.app.Activity
import android.os.Bundle
import android.widget.LinearLayout
import android.widget.TextView

class MainActivity : Activity() {
    // onCreate 构造演示任务详情界面，用于验证眼镜端节点展示和必填标记。
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val task = InspectionTask(
            id = 1,
            title = "机房日巡检 #A-102",
            dueAt = "09:30",
            nodes = listOf(
                InspectionNode(1, "到达 A 区设备柜", 0, false, false),
                InspectionNode(2, "拍摄设备面板状态", 1, false, true),
                InspectionNode(3, "记录温度与指示灯", 1, true, true)
            )
        )
        val layout = LinearLayout(this)
        layout.orientation = LinearLayout.VERTICAL
        layout.setPadding(32, 32, 32, 32)
        layout.addView(TextView(this).apply { text = "我的任务"; textSize = 22f })
        layout.addView(TextView(this).apply { text = "${task.title} · 截止 ${task.dueAt}"; textSize = 18f })
        task.nodes.forEach { node ->
            layout.addView(TextView(this).apply {
                text = "${node.id}. ${node.name} ${if (node.minPhotos > 0) "· 必拍照" else ""} ${if (node.requireText) "· 必填" else ""}"
                textSize = 16f
            })
        }
        layout.addView(TextView(this).apply { text = "操作：拍照 / 语音备注 / 上报异常 / 提交"; textSize = 16f })
        setContentView(layout)
    }
}
