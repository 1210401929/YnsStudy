<template>
  <el-aside width="100%" class="sidebar">
    <div class="sidebar-header">
      <span class="sidebar-title">{{ view_title }}</span>
      <el-button v-if="isSelf()" type="primary" link class="add-btn" :icon="Plus" @click="openCategoryDialog(null)">
        新增分类
      </el-button>
    </div>

    <div class="menu-scroll-container">
      <el-menu
          class="custom-el-menu"
          :default-active="selectedIndex"
          :default-openeds="defaultOpeneds"
          @select="handleSelect">

        <el-sub-menu
            v-for="cat1 in categoryMenuTree"
            :key="cat1.GUID"
            :index="'cat-' + cat1.GUID"
            @contextmenu.stop.prevent="handleRightClick($event, 'category', cat1.GUID, 1)">

          <template #title>
            <el-icon>
              <Folder/>
            </el-icon>
            <span class="menu-text">{{ cat1.CATNAME }}</span>
          </template>

          <el-sub-menu
              v-for="cat2 in cat1.children"
              :key="cat2.GUID"
              :index="'cat-' + cat2.GUID"
              @contextmenu.stop.prevent="handleRightClick($event, 'category', cat2.GUID, 2)">

            <template #title>
              <el-icon>
                <FolderOpened/>
              </el-icon>
              <span class="menu-text">{{ cat2.CATNAME }}</span>
            </template>

            <el-menu-item
                v-for="blog in getBlogsByCategory(cat2.GUID)"
                :key="blog.GUID"
                :index="blog.GUID"
                class="el-menu_"
                @click="clickBlog(blog)"
                @contextmenu.stop.prevent="handleRightClick($event, 'blog', blog.GUID)">
              <el-tag :type="blog.BLOG_TYPE === 'public' ? 'success' : 'info'" effect="light" size="small"
                      class="custom-tag">
                {{ blog.BLOG_TYPE === "public" ? "公" : "私" }}
              </el-tag>
              <span class="blog-title-text" :title="blog.BLOG_TITLE">{{ blog.BLOG_TITLE }}</span>
            </el-menu-item>
          </el-sub-menu>

          <el-menu-item
              v-for="blog in getBlogsByCategory(cat1.GUID)"
              :key="blog.GUID"
              :index="blog.GUID"
              class="el-menu_"
              @click="clickBlog(blog)"
              @contextmenu.stop.prevent="handleRightClick($event, 'blog', blog.GUID)">
            <el-tag :type="blog.BLOG_TYPE === 'public' ? 'success' : 'info'" effect="light" size="small"
                    class="custom-tag">
              {{ blog.BLOG_TYPE === "public" ? "公" : "私" }}
            </el-tag>
            <span class="blog-title-text" :title="blog.BLOG_TITLE">{{ blog.BLOG_TITLE }}</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="cat-default">
          <template #title>
            <el-icon>
              <Document/>
            </el-icon>
            <span class="menu-text">未归档</span>
          </template>
          <el-menu-item
              v-for="blog in getBlogsByCategory(null)"
              :key="blog.GUID"
              :index="blog.GUID"
              class="el-menu_"
              @click="clickBlog(blog)"
              @contextmenu.stop.prevent="handleRightClick($event, 'blog', blog.GUID)">
            <el-tag :type="blog.BLOG_TYPE === 'public' ? 'success' : 'info'" effect="light" size="small"
                    class="custom-tag">
              {{ blog.BLOG_TYPE === "public" ? "公" : "私" }}
            </el-tag>
            <span class="blog-title-text" :title="blog.BLOG_TITLE">{{ blog.BLOG_TITLE }}</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </div>
  </el-aside>

  <transition v-if="isSelf()" name="el-zoom-in-top">
    <div v-show="contextMenu.visible" class="custom-context-menu"
         :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }">
      <template v-if="contextMenu.type === 'category'">
        <div v-if="contextMenu.level === 1" class="menu-item" @click="openCategoryDialog(contextMenu.targetId)">
          <el-icon>
            <FolderAdd/>
          </el-icon>
          新增下级分类
        </div>
        <div class="menu-item" @click="updateCategory(contextMenu)">
          <el-icon>
            <EditPen/>
          </el-icon>
          修改分类
        </div>
        <div class="divider"></div>
        <div class="menu-item danger" @click="handleDeleteCategory(contextMenu.targetId)">
          <el-icon>
            <Delete/>
          </el-icon>
          删除分类
        </div>
      </template>
      <template v-if="contextMenu.type === 'blog'">
        <div class="menu-item" @click="openMoveBlogDialog">
          <el-icon>
            <Switch/>
          </el-icon>
          修改分类
        </div>
      </template>
    </div>
  </transition>

  <el-dialog v-model="categoryDialogVisible" :title="currentParentId ? '新增下级分类' : '新增一级分类'" width="450px"
             append-to-body>
    <el-form label-width="85px" label-position="right" @submit.prevent style="margin-top: 10px;">
      <el-form-item label="分类名称" required>
        <el-input v-model="newCategoryName" placeholder="请输入分类名称" maxlength="20" show-word-limit
                  @keyup.enter="submitNewCategory"/>
      </el-form-item>
      <el-form-item label="分类描述">
        <el-input v-model="newCategoryRemark" type="textarea" :rows="3" placeholder="请输入描述信息（可选）"
                  maxlength="100" show-word-limit resize="none"/>
      </el-form-item>
      <el-form-item label="排序序号">
        <div style="display: flex; align-items: center; gap: 12px; width: 100%;">
          <el-input v-model="newCategoryOrderNo" style="width: 130px;"/>
          <span style="color: #909399; font-size: 12px; line-height: 1;">数字越小越靠前</span>
        </div>
      </el-form-item>
    </el-form>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitNewCategory">确定</el-button>
      </span>
    </template>
  </el-dialog>

  <el-dialog v-model="moveBlogDialogVisible" title="修改文章分类" width="400px" append-to-body>
    <div style="margin-bottom: 15px; color: #606266;">请选择目标分类：</div>
    <el-tree-select
        v-model="targetCategoryId"
        :data="categoryTreeData"
        check-strictly
        :render-after-expand="false"
        placeholder="选择分类（留空则为未归档）"
        style="width: 100%"
        clearable
    />
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="moveBlogDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitMoveBlog">确定</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import {ref, computed, onMounted, onBeforeUnmount, reactive, watch} from 'vue'
import {useRouter} from 'vue-router'
import {useBlogContentStore} from '@/stores/detail/blog.js'
import {Folder, FolderOpened, FolderAdd, EditPen, Delete, Document, Plus, Switch} from '@element-plus/icons-vue'
import {ElMessage} from 'element-plus'
import {ele_confirm, getGuid, sendAxiosRequest} from "@/utils/common.js";


const props = defineProps({
  view_title: {type: String, default: '归档目录'},
  selectedIndex: {type: String, default: ''},
  userCode: {type: [String, Number],default: ''},
  isRouter: {type: Boolean, default: false}
})

const emit = defineEmits(['update:selectedIndex', 'clickBlog'])
const router = useRouter()
const blogContentStore = useBlogContentStore()

//是否显示操作相关内容
const isSelf = () => {
  //暂时判断，如果需要操控路由，就表示是自己的分类目录  因为是在发表模块调用的
  return props.isRouter;
}

// ================= 你的所有逻辑代码保持完全一致 =================
const categoryList = ref([])

function getPageData() {
  const getUserBlogCat = async () => {
    let result = await sendAxiosRequest("/blog-api/blog/getUserBlogCat", {userCode: props.userCode});
    categoryList.value = result.result;
  }
  getUserBlogCat();
}

const categoryMenuTree = computed(() => {
  const topLevel = categoryList.value.filter(c => !c.SUPERGUID || c.SUPERGUID === '0' || c.SUPERGUID === 'null');
  return topLevel.map(cat1 => {
    const children = categoryList.value.filter(c => c.SUPERGUID === cat1.GUID);
    return {...cat1, children: children}
  });
});

const getBlogsByCategory = (categoryId) => {
  return blogContentStore.blogContents.filter(blog => {
    if (categoryId) return blog.CAT_ID === categoryId;
    return !blog.CAT_ID;
  })
}

const defaultOpeneds = ref([])
let isDefaultOpenedSet = false

watch(
    [() => blogContentStore.blogContents, categoryMenuTree],
    ([blogs, tree]) => {
      if (isDefaultOpenedSet && tree.length === 0) return;
      const unarchivedBlogs = blogs.filter(b => !b.CAT_ID);

      if (unarchivedBlogs.length > 0) {
        defaultOpeneds.value = ['cat-default'];
        handleSelect(unarchivedBlogs[0].GUID);
        isDefaultOpenedSet = true;
        return;
      }

      if (tree.length > 0 && blogs.length > 0) {
        for (const cat1 of tree) {
          const cat1Blogs = blogs.filter(b => b.CAT_ID === cat1.GUID);
          if (cat1Blogs.length > 0) {
            defaultOpeneds.value = ['cat-' + cat1.GUID];
            handleSelect(cat1Blogs[0].GUID);
            isDefaultOpenedSet = true;
            return;
          }
          if (cat1.children && cat1.children.length > 0) {
            for (const cat2 of cat1.children) {
              const cat2Blogs = blogs.filter(b => b.CAT_ID === cat2.GUID);
              if (cat2Blogs.length > 0) {
                defaultOpeneds.value = ['cat-' + cat1.GUID, 'cat-' + cat2.GUID];
                handleSelect(cat2Blogs[0].GUID);
                isDefaultOpenedSet = true;
                return;
              }
            }
          }
        }
      }
    },
    {deep: true, immediate: true}
)

const categoryTreeData = computed(() => {
  return categoryMenuTree.value.map(cat1 => {
    return {
      value: cat1.GUID,
      label: cat1.CATNAME,
      children: cat1.children && cat1.children.length > 0
          ? cat1.children.map(cat2 => ({value: cat2.GUID, label: cat2.CATNAME}))
          : undefined
    }
  });
});

const contextMenu = reactive({visible: false, x: 0, y: 0, type: '', targetId: null, level: 1})

const handleRightClick = (e, type, id, level = 1) => {
  contextMenu.type = type;
  contextMenu.targetId = id;
  contextMenu.level = level;
  contextMenu.visible = true;
  contextMenu.x = e.clientX + 10;
  contextMenu.y = e.clientY + 10;
}

const closeContextMenu = () => contextMenu.visible = false;

const categoryDialogVisible = ref(false)
const newCategoryName = ref('')
const newCategoryRemark = ref('')
const newCategoryOrderNo = ref(1)
const currentParentId = ref(null)
const editingCategoryId = ref(null)

const openCategoryDialog = (parentId = null) => {
  editingCategoryId.value = null;
  currentParentId.value = parentId;
  newCategoryName.value = '';
  newCategoryRemark.value = '';

  const existingChildrenCount = categoryList.value.filter(c =>
      parentId ? c.SUPERGUID === parentId : (!c.SUPERGUID || c.SUPERGUID === '0' || c.SUPERGUID === 'null')
  ).length;

  newCategoryOrderNo.value = existingChildrenCount + 1;
  categoryDialogVisible.value = true;
  closeContextMenu();
}

const updateCategory = (menu) => {
  const targetCat = categoryList.value.find(c => c.GUID === menu.targetId);
  if (targetCat) {
    editingCategoryId.value = targetCat.GUID;
    currentParentId.value = targetCat.SUPERGUID;

    newCategoryName.value = targetCat.CATNAME || '';
    newCategoryRemark.value = targetCat.REMARK || '';
    newCategoryOrderNo.value = targetCat.ORDERNO || 1;

    categoryDialogVisible.value = true;
  }
  closeContextMenu();
}

const handleDeleteCategory = (categoryId) => {
  ele_confirm(`该操作会删除分类下所有子分类，并移出相关文章到未归档中`, async () => {
    let result = await sendAxiosRequest("/blog-api/blog/deleteBlogCat", {guid: categoryId});
    if (result.isError) return ElMessage.error("删除失败");

    categoryList.value = categoryList.value.filter(cat => {
      return cat.GUID !== categoryId && cat.SUPERGUID !== categoryId;
    });

    ElMessage.success('分类删除成功！');
    blogContentStore.blogContents.forEach(blog => {
      if (blog.CAT_ID === categoryId) {
        blog.CAT_ID = null;
      }
    });
  })
}

const submitNewCategory = async () => {
  if (!newCategoryName.value.trim()) return ElMessage.warning('分类名称不能为空');

  if (editingCategoryId.value) {
    let updateData = {
      GUID: editingCategoryId.value,
      CATNAME: newCategoryName.value.trim(),
      REMARK: newCategoryRemark.value.trim(),
      ORDERNO: newCategoryOrderNo.value
    };

    let result = await sendAxiosRequest("/blog-api/blog/updateBlogCat", {blogCat: updateData});
    if (!result || result.isError) {
      ElMessage.error(result.errMsg);
    }

    const index = categoryList.value.findIndex(c => c.GUID === editingCategoryId.value);
    if (index !== -1) {
      categoryList.value[index].CATNAME = updateData.CATNAME;
      categoryList.value[index].REMARK = updateData.REMARK;
      categoryList.value[index].ORDERNO = updateData.ORDERNO;
    }
    ElMessage.success('分类修改成功！');
  } else {
    let blogCat = {
      GUID: getGuid(),
      USERCODE: props.userCode,
      CATNAME: newCategoryName.value.trim(),
      SUPERGUID: currentParentId.value,
      REMARK: newCategoryRemark.value.trim(),
      ORDERNO: newCategoryOrderNo.value
    }

    let result = await sendAxiosRequest("/blog-api/blog/addBlogCat", {blogCat});
    if (!result || result.isError) {
      ElMessage.error(result.errMsg);
    }
    categoryList.value.push(blogCat);
    ElMessage.success('分类创建成功！');
  }

  categoryList.value.sort((a, b) => a.ORDERNO - b.ORDERNO);
  categoryDialogVisible.value = false;
}

const moveBlogDialogVisible = ref(false)
const targetCategoryId = ref(null)

const openMoveBlogDialog = () => {
  const targetBlog = blogContentStore.blogContents.find(b => b.GUID === contextMenu.targetId);
  if (targetBlog) targetCategoryId.value = targetBlog.CAT_ID || null;
  moveBlogDialogVisible.value = true;
  closeContextMenu();
}

const submitMoveBlog = () => {
  const blog = blogContentStore.blogContents.find(b => b.GUID === contextMenu.targetId);
  if (blog) {
    blog.CAT_ID = targetCategoryId.value || null;
    ElMessage.success('修改分类成功！');
  }
  sendAxiosRequest("/blog-api/blog/updateBlogCatId", {blogId: blog.GUID, catId: blog.CAT_ID});
  moveBlogDialogVisible.value = false;
}

function handleSelect(index) {
  const item = blogContentStore.blogContents.find(b => b.GUID === index)
  if (item) {
    emit('update:selectedIndex', index)
    //如果需要跳转路由：发表模块
    if (props.isRouter)
      router.push({name: 'BlogContent', query: {g: item.GUID}})
  }
}

const clickBlog = (blogInfo) => {
  emit('clickBlog', blogInfo)
}

onMounted(() => {
  if (!blogContentStore.blogContents || blogContentStore.blogContents.length === 0)
    blogContentStore.initBlogContent(props.userCode);
  window.addEventListener('click', closeContextMenu);
  getPageData();
})
onBeforeUnmount(() => window.removeEventListener('click', closeContextMenu))
</script>

<style scoped>
/* ================= 核心容器美化 (纯透明，无模糊) ================= */
.sidebar {
  background-color: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.04);

  height: calc(100vh - 100px);
  max-height: 800px;
  position: sticky;
  top: 20px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: all 0.3s ease;
}

/* ================= 头部美化 ================= */
.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.04);
  /* 头部背景也设置为透明 */
  background: transparent;
  z-index: 10;
}

.sidebar-title {
  font-weight: 600;
  font-size: 15px;
  color: #2c3e50;
  letter-spacing: 0.5px;
}

.add-btn {
  font-weight: 500;
}

/* ================= 滚动区域美化 (定制极细滚动条) ================= */
.menu-scroll-container {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 8px 0 16px 0;
}

.menu-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.menu-scroll-container::-webkit-scrollbar-track {
  background: transparent;
}

.menu-scroll-container::-webkit-scrollbar-thumb {
  background: rgba(144, 147, 153, 0.2);
  border-radius: 4px;
}

.menu-scroll-container:hover::-webkit-scrollbar-thumb {
  background: rgba(144, 147, 153, 0.4);
}

/* ================= Element-plus 菜单深度复写 ================= */
/* 核心修复：强制所有层级的菜单及下拉面板背景纯透明 */
:deep(.el-menu),
:deep(.el-menu--inline),
.custom-el-menu {
  background-color: transparent !important;
  border-right: none !important;
}

/* 压缩菜单项高度，制作胶囊形状，并确保默认背景透明 */
:deep(.el-sub-menu__title),
:deep(.el-menu-item) {
  background-color: transparent !important;
  height: 38px !important;
  line-height: 38px !important;
  margin: 4px 12px;
  border-radius: 8px;
  color: #606266;
  transition: all 0.25s ease;
}

:deep(.el-menu-item) {
  padding-left: 44px !important;
}

/* Hover状态微透出主题色 */
:deep(.el-sub-menu__title:hover),
:deep(.el-menu-item:hover) {
  background-color: rgba(64, 158, 255, 0.08) !important;
  color: #409eff;
}

/* 选中状态 */
:deep(.el-menu-item.is-active) {
  background-color: rgba(64, 158, 255, 0.1) !important;
  color: #409eff;
  font-weight: 600;
}

/* 菜单文本和图标对齐 */
.menu-text {
  font-weight: 500;
  font-size: 14px;
}

.el-menu_ {
  display: flex;
  align-items: center;
  gap: 6px;
}

.blog-title-text {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  flex: 1;
  font-size: 13px;
}

.custom-tag {
  transform: scale(0.85);
  transform-origin: left center;
  border-radius: 4px;
  border: none;
}

/* ================= 右键菜单纯透明美化 ================= */
.custom-context-menu {
  position: fixed;
  z-index: 2000;
  background-color: rgba(255, 255, 255, 0.95); /* 稍微高一点透明度保证文字清晰，无模糊 */
  border: 1px solid rgba(228, 231, 237, 0.8);
  border-radius: 10px;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.12);
  padding: 6px;
  min-width: 140px;
}

.custom-context-menu .menu-item {
  padding: 8px 12px;
  margin: 2px 0;
  font-size: 13px;
  color: #606266;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  border-radius: 6px;
  transition: background-color 0.2s, color 0.2s;
  background-color: transparent;
}

.custom-context-menu .menu-item:hover {
  background-color: rgba(64, 158, 255, 0.1);
  color: #409eff;
}

.custom-context-menu .menu-item.danger:hover {
  background-color: rgba(245, 108, 108, 0.1);
  color: #f56c6c;
}

.custom-context-menu .divider {
  height: 1px;
  background-color: rgba(0, 0, 0, 0.06);
  margin: 4px 0;
}

/* 响应式 */
@media screen and (max-width: 768px) {
  .sidebar {
    width: 100%;
    margin-top: 0;
    position: static;
    height: auto;
    max-height: 500px;
  }
}
</style>