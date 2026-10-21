// 便签墙的编辑、删除权限，与 Go 后端一致（前端只决定显示哪些按钮，真正的限制在后端）
import { adminUserCode } from '@/config/vue-config.js'
import { getCurrentUserAdminObject } from '@/utils/common.js'

// 编辑：自己的可以改；管理员可以改除超级管理员以外所有人的；超级管理员都能改
export function canEditNote(note, userBean) {
    if (!note) return false
    const admin = getCurrentUserAdminObject()
    if (admin.adminLevel === 'superAdmin') return true
    if (userBean?.code && userBean.code === note.USERCODE) return true
    return !!admin.isAdmin && note.USERCODE !== adminUserCode
}

// 删除：与编辑相同，另外置顶的随记只有超级管理员能删
export function canDeleteNote(note, userBean) {
    if (!note) return false
    if (getCurrentUserAdminObject().adminLevel === 'superAdmin') return true
    if (note.ISTOP === '1') return false
    return canEditNote(note, userBean)
}
