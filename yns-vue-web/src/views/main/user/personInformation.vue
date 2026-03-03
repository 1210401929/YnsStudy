<template>
  <div class="person-info" :style="bgStyle">
    <!-- 左侧悬浮设置面板（不占文档流） -->
    <div class="left-dock"
         v-if="isSelf"
         :class="{ collapsed: dockCollapsed }">
      <el-card shadow="hover" class="dock-card">
        <div class="dock-title">🎨 背景设置</div>

        <!-- 背景图片 -->
        <div class="dock-block">
          <div class="dock-subtitle">图片</div>
          <el-input
              v-model="bgImageInput"
              size="small"
              placeholder="粘贴图片URL后回车"
              @keyup.enter="applyBgImageUrl"
              style="margin-top:8px"/>
          <div class="dock-actions">
            <el-button size="small" @click="resetBgImage">恢复默认</el-button>
            <el-button size="small" link @click="previewBgImage" :disabled="!bgImage">预览</el-button>
          </div>
        </div>

        <el-divider/>

        <!-- 背景音乐 -->
        <div class="dock-title">
          {{ isSelf ? "🎵 音乐设置" : `${user.name}的背景音乐` }}
        </div>
        <div class="dock-block">
          <el-input
              v-model="bgAudioInput"
              size="small"
              placeholder="粘贴音频URL后回车"
              @keyup.enter="applyBgAudioUrl"
              style="margin-top:8px"/>

          <div class="audio-row">
            <el-button size="small" :type="isAudioPlaying ? 'danger' : 'primary'" @click="toggleAudio">
              {{ isAudioPlaying ? '暂停' : '播放' }}
            </el-button>
            <el-switch
                v-model="audioMuted"
                size="small"
                active-text="静音"
                style="margin-left:12px"
                @change="applyMute"/>
          </div>

          <div class="audio-row" style="margin-top:8px">
            <span class="vol-label">音量</span>
            <el-slider v-model="audioVolume" :min="0" :max="100" :step="1" style="flex:1" @change="applyVolume"/>
          </div>
        </div>
      </el-card>

      <!-- 折叠按钮 -->
      <button class="dock-toggle" @click="dockCollapsed = !dockCollapsed">
        {{ dockCollapsed ? '▶' : '◀' }}
      </button>
    </div>
    <!-- 隐藏原生控件的播放器 -->
    <audio
        ref="bgAudioRef"
        :src="bgAudio || ''"
        loop
        preload="auto"
        playsinline
    ></audio>
    <!-- 访客迷你音乐控制（仅当不是本人 & 有音乐地址时出现） -->
    <div class="music-mini" v-if="!isSelf && bgAudio">
      <div class="music-mini__title">🎵 {{ user.name || 'Ta' }} 的音乐</div>
      <div class="music-mini__row mini-controls">
        <!-- 播放 / 暂停 -->
        <el-button
            class="icon-btn"
            circle
            size="small"
            :title="isAudioPlaying ? '暂停' : '播放'"
            :type="isAudioPlaying ? 'danger' : 'primary'"
            @click="toggleAudio"
        >
          {{ isAudioPlaying ? '⏸' : '▶' }}
        </el-button>

        <!-- 静音切换 -->
        <el-button
            class="icon-btn"
            circle
            size="small"
            :title="audioMuted ? '取消静音' : '静音'"
            @click="audioMuted = !audioMuted; applyMute()"
        >
          {{ audioMuted ? '🔇' : '🔊' }}
        </el-button>

        <!-- 精简音量条 -->
        <el-slider
            class="mini-slider"
            v-model="audioVolume"
            :min="0"
            :max="100"
            :step="1"
            :show-tooltip="false"
            @input="applyVolume"
        />
      </div>
    </div>


    <div class="inner-container">
      <!-- 顶部跳转按钮 -->
      <div class="toolbar">
        <el-button type="success" size="small" @click="scrollToFiles">
          📁 跳转到文件列表
        </el-button>
      </div>

      <!-- 用户信息卡片 -->
      <el-card class="profile-card">
        <div class="profile-header">
          <!-- 左侧：头像和基础信息 -->
          <div class="left-info">
            <el-avatar
                :src="user.avatar"
                size="large"
                class="author-avatar"
                alt="用户头像"
            >
              {{ user.name?.charAt(0) }}
            </el-avatar>
            <div class="profile-details">
              <h2 class="username">
                <span class="name">{{ user.name }}</span>
                <span v-if="user.code === adminUserCode" class="public-badge superAdmin-badge">超级管理员</span>
                <span v-if="user.role === 'admin'" class="public-badge admin-badge">管理员</span>
                <span v-if="user.isban === '1'" class="public-badge ban-badge">已封禁</span>
              </h2>
              <p>邮箱: {{ user.email || "未知" }}</p>
              <p class="user-remark" :title="user.remark">个性签名: {{ user.remark || "这个人很神秘~" }}</p>
              <p class="userip">IP: {{ user.loginaddress || "未知" }}</p>
            </div>
          </div>

          <!-- 右侧：互动信息 -->
          <div class="right-info">
            <div class="stats">
              <span @click="followingUserClick" style="cursor: pointer;">关注 <strong>{{ followingNum }}</strong></span>
              <span @click="followersUserClick" style="cursor: pointer;">粉丝 <strong>{{ followersNum }}</strong></span>
            </div>
            <div class="author-actions">
              <el-button
                  :type="isFollowing ? 'danger' : 'primary'"
                  size="small"
                  round
                  @click="toggleFollow"
                  class="follow-button"
              >
                {{ isFollowing ? '取消关注' : '关注' }}
              </el-button>
              <el-button size="small" class="follow-button" round @click="messageAuthor">私信</el-button>
            </div>
            <div class="interaction-buttons">
              <el-button type="primary" link @click="goToLiked">👍 点赞数:{{ likeNum }}</el-button>
              <el-button type="warning" link @click="goToFavorites">⭐ 收藏数:{{ collectNum }}</el-button>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 博客内容 -->
      <el-card class="section-card">
        <div class="section-header">
          <h3 class="section-title">📚 发表内容</h3>
          <el-switch
              v-model="onlyArticle"
              size="small"
              active-text="仅文章"
              style="margin-left: auto;"
          />
        </div>
        <el-timeline v-if="showBlogs.length>0">
          <el-timeline-item
              v-for="(blog, index) in showBlogs"
              :key="blog.GUID"
              :timestamp="(blog.TYPE=='blog' ? '发表文章：' : '社区发起：') + formatDate(blog.CREATE_TIME)"
              placement="top"
              type="primary"
          >
            <el-card shadow="hover" class="blog-card" @click="blogMainClick(blog)">
              <h4>{{ blog.BLOG_TITLE }}</h4>
              <p class="blog-summary" v-html="stripImages(blog.MAINTEXT)"></p>
            </el-card>
          </el-timeline-item>
        </el-timeline>

        <el-empty v-if="!showBlogs.length && !loading" description="暂无内容"/>
        <el-button
            v-if="!noMore && !loading"
            type="primary"
            link
            @click="fetchArticles(null)"
            style="margin: 20px auto; display: block;"
        >
          加载更多
        </el-button>
        <div v-if="loading" class="loading-text">加载中...</div>
        <div v-if="noMore" class="end-text">没有更多内容了</div>
      </el-card>

      <!-- 上传的文件 -->
      <el-card class="section-card" ref="fileSection">
        <h3 class="section-title">📁 上传的文件</h3>
        <el-empty v-if="files.length === 0" description="暂无上传文件"/>
        <el-table
            v-else
            :data="files"
            stripe
            border
            style="width: 100%"
            class="file-table"
        >
          <el-table-column prop="ORIGINALFILENAME" label="文件名"/>
          <el-table-column prop="DOWNNUM" label="下载次数" width="100"/>
          <el-table-column
              prop="CREATE_TIME"
              label="上传时间"
              :formatter="(row) => formatDate(row.CREATE_TIME)"
          />
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button type="primary" link @click="downloadFile(scope.row)">下载</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
    <el-dialog
        v-model="showDialog"
        width="80%"
        destroy-on-close
        top="4vh"
        @close="onDialogClose"
    >
      <ContentAndComment :blogId="selectedBlogId"/>
    </el-dialog>
    <!-- 通用卡片结构（点赞/收藏列表都用） -->
    <el-dialog v-model="showLikeDialog" title="👍 点赞列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
      <div class="card-list-scroll">
        <div class="card-list">
          <!-- 卡片内容保持统一结构 -->
          <el-card
              v-for="item in likeList"
              :key="item.GUID"
              class="list-card"
              shadow="hover"
              @click="blogMainClick(item)"
          >
            <div class="card-top">
              <span class="author-name" style="color: #0029fc">作者:{{ item.USERNAME }}</span>
              <span class="card-time">{{ formatDate(item.CREATE_TIME) }}</span>
            </div>
            <div class="card-body">
              <h4 class="card-title">{{ item.BLOG_TITLE }}</h4>
            </div>
          </el-card>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="showCollectDialog" title="⭐ 收藏列表" width="60%" :modal-class="'fixed-dialog-height'"
               top="6vh">
      <div class="card-list-scroll">
        <div class="card-list">
          <!-- 卡片内容保持统一结构 -->
          <el-card
              v-for="item in collectList"
              :key="item.GUID"
              class="list-card"
              shadow="hover"
              @click="blogMainClick(item)"
          >
            <div class="card-top">
              <span class="author-name" style="color: #0029fc">作者:{{ item.USERNAME }}</span>
              <span class="card-time">{{ formatDate(item.CREATE_TIME) }}</span>
            </div>
            <div class="card-body">
              <h4 class="card-title">{{ item.BLOG_TITLE }}</h4>
            </div>
          </el-card>

        </div>
      </div>
    </el-dialog>
    <!-- 粉丝列表-->
    <el-dialog v-model="showFollowersUser" title="🙋粉丝列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
      <div class="card-list-scroll">
        <div class="card-list">
          <!-- 卡片内容保持统一结构 -->
          <el-card
              v-for="item in followersUser"
              :key="item.CODE"
              class="list-card-follow"
              shadow="hover"
              @click="openFollowersUser(item)"
          >
            <div class="follower-card-content">
              <el-avatar
                  :src="item.avatar"
                  size="large"
                  class="author-avatar-follow"
                  alt="用户头像"
              >
                {{ item.NAME?.charAt(0) }}
              </el-avatar>
              <span class="author-name">{{ item.NAME }}</span>
            </div>
          </el-card>
        </div>
      </div>
    </el-dialog>
    <!-- 关注列表-->
    <el-dialog v-model="showFollowingUser" title="👀关注列表" width="60%" :modal-class="'fixed-dialog-height'" top="6vh">
      <div class="card-list-scroll">
        <div class="card-list">
          <!-- 卡片内容保持统一结构 -->
          <el-card
              v-for="item in followingUser"
              :key="item.CODE"
              class="list-card-follow"
              shadow="hover"
              @click="openFollowingUser(item)"
          >
            <div class="follower-card-content">
              <el-avatar
                  :src="item.avatar"
                  size="large"
                  class="author-avatar-follow"
                  alt="用户头像"
              >
                {{ item.NAME?.charAt(0) }}
              </el-avatar>
              <span class="author-name">{{ item.NAME }}</span>
            </div>
          </el-card>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import {ref, onMounted, computed,nextTick} from 'vue'
import {ElMessage} from 'element-plus'
import {useRoute, useRouter} from 'vue-router'
import {
  pubFormatDate,
  decrypt,
  sendAxiosRequest,
  stripImages,
  downloadFileByUrl,
  sendNotifications, extractPlainTextFromHTML, getUserInfoByCode
} from '@/utils/common.js'
import ContentAndComment from '@/views/detail/blog/ContentAndComment.vue'
import {adminUserCode} from '@/config/vue-config.js'
import {useUserStore} from '@/stores/main/user.js'
import {pubOpenUser} from "@/utils/blogUtil.js";

const userStore = useUserStore()
userStore.initFromLocal()

const showLikeDialog = ref(false)
const showCollectDialog = ref(false)
const likeNum = ref(0)
const likeList = ref([])
const collectNum = ref(0)
const collectList = ref([])

const route = useRoute()
const router = useRouter()
const user = ref({})

//当前打开用户code
const targetUserCode = ref('');
try {
  const raw = route.params.u;
  targetUserCode.value = raw ? decrypt(raw) : '';
} catch (e) {
  targetUserCode.value = '';
}

// 是否访问自己的主页
const isSelf = computed(() =>
    !!userStore.userBean.code &&
    !!targetUserCode.value &&
    targetUserCode.value === userStore.userBean.code
);
// 关注相关
const isFollowing = ref(false)
const followersNum = ref(0)
const followersUser = ref([])
const followingNum = ref(0)
const followingUser = ref([])

const showFollowersUser = ref(false)
const showFollowingUser = ref(false)
const toggleFollow = () => {
  if (!userStore.userBean.code) {
    ElMessage.error('用户过期,请返回主页面重新登录!')
    return false
  }
  isFollowing.value = !isFollowing.value
  if (isFollowing.value) {
    followersNum.value++
    sendAxiosRequest('/blog-api/userInformation/followUser', {
      followUserCode: user.value.code,
      followUserName: user.value.name
    })
    //发送消息
    sendNotifications(userStore.userBean.code, user.value.code, "followUser", null, `${userStore.userBean.name}关注了你`)

  } else {
    followersNum.value--
    sendAxiosRequest('/blog-api/userInformation/noFollowUser', {followUserCode: user.value.code})
  }
  ElMessage.success(isFollowing.value ? '已关注' : '已取消关注')
}

const followingUserClick = () => {
  if (followingUser.value.length === 0) {
    ElMessage.success('他很高冷,没有关注任何人')
    return false
  }
  showFollowingUser.value = true
}

const followersUserClick = () => {
  if (followersUser.value.length === 0) {
    ElMessage.success('他没有粉丝,仍需努力')
    return false
  }
  showFollowersUser.value = true
}

const openFollowersUser = (item) => {
  pubOpenUser(router, item.CODE);
}

const openFollowingUser = (item) => {
  pubOpenUser(router, item.CODE);
}

const page = ref(1)
const pageSize = 5
const loading = ref(false)
const noMore = ref(false)

const fetchArticles = async (userCode) => {
  if (loading.value || noMore.value) return;
  if (!userCode) userCode = decrypt(route.params.u);
  loading.value = true
  try {
    const res = await sendAxiosRequest('/blog-api/userInformation/getBlogAndCommunityByUserCode', {
      userCode,
      page: page.value,
      pageSize,
      keyword: ""
    })
    const newData = res.result.data
    if (newData.length < pageSize) {
      noMore.value = true
    }
    blogs.value.push(...newData)
    page.value++
  } catch (e) {
    console.error('获取内容失败', e)
  } finally {
    loading.value = false
  }
}

async function getUserInfo2Data() {
  let userCode = route.params.u;
  if (userCode) {
    userCode = decrypt(userCode)
    //获取用户发布内容  文章和社区
    fetchArticles(userCode);
    //获取用户信息
    let result = await getUserInfoByCode(userCode);
    if (result && !result.isError) {
      user.value = result.result
    }
    //获取用户上传的文件
    result = await sendAxiosRequest('/blog-api/userInformation/getResourceByUserCode', {userCode})
    if (result && !result.isError) {
      files.value = result.result;
    }
    //获取用户点赞作品和收藏作品
    result = await sendAxiosRequest('/blog-api/blog/getLikeAndCollectByUserCode', {userCode})
    if (result && !result.isError) {
      result.result.forEach(item => {
        if (item.TYPE === 'like') {
          likeNum.value++
          likeList.value.push(item)
        } else if (item.TYPE === 'collect') {
          collectNum.value++
          collectList.value.push(item)
        }
      })
    }
    //获取用户的关注列表和粉丝列表
    result = await sendAxiosRequest('/blog-api/userInformation/getFollowUser', {userCode})
    if (result && !result.isError) {
      followersNum.value = result.result.followersUser.length
      followersUser.value = result.result.followersUser
      followingNum.value = result.result.followingUser.length
      followingUser.value = result.result.followingUser
      const arr = result.result.followersUser.filter(item => item.CODE === userStore.userBean.code)
      if (arr.length > 0) isFollowing.value = true
    }
    //设置用户主页信息  :背景图片,背景音乐等等
    setPersonInfo();
  }
  //修改浏览器title和meta,有助于搜索排名
  document.title = user.value.name + "的主页";
  const metaDesc = document.querySelector('meta[name="description"]');
  if (metaDesc) {
    metaDesc.setAttribute("content", document.title);
  } else {
    const desc = document.createElement('meta')
    desc.name = "description";
    desc.content = document.title;
    document.head.appendChild(desc)
  }
}

getUserInfo2Data()

const blogs = ref([])
const files = ref([])
const onlyArticle = ref(false)
const showDialog = ref(false)
const selectedBlogId = ref('')

const showBlogs = computed(() => {
  if (!onlyArticle.value)
    return blogs.value;
  else {
    return blogs.value.filter(item => item.TYPE == "blog");
  }
})

function blogMainClick(blog) {
  if (blog.TYPE == "community") {
    ElMessage.success('社区发起内容不用查看详情')
    return false
  }
  selectedBlogId.value = blog.GUID
  showDialog.value = true
}

const onDialogClose = () => {
  selectedBlogId.value = ''
  showDialog.value = false
}

const formatDate = (dateStr) => pubFormatDate(dateStr)

const downloadFile = (file) => {
  ElMessage.success(`准备下载：${file.ORIGINALFILENAME}`)
  downloadFileByUrl(file.FILEVIEWURL, file.ORIGINALFILENAME)
}

const fileSection = ref(null)
const scrollToFiles = () => {
  if (fileSection.value?.$el) {
    fileSection.value.$el.scrollIntoView({behavior: 'smooth'})
  }
}

const goToLiked = () => {
  if (likeList.value.length === 0) {
    ElMessage.success('他没有点赞的文章,也没有喜欢的人')
    return false
  }
  showLikeDialog.value = true
}

const goToFavorites = () => {
  if (collectList.value.length === 0) {
    ElMessage.success('他没有收藏的文章,也没有喜欢的人')
    return false
  }
  showCollectDialog.value = true
}

const messageAuthor = () => {
  ElMessage.success('私信作者:' + user.value.name)
}

// ==== 背景与音乐：状态 ====
const defaultBg = ""; // 你的默认背景
const bgImage = ref("");
const bgImageInput = ref("");

const bgPresets = {
  // 贴近你默认的天空蓝 → 更亮的白，带两处很淡的蓝色光斑
  softSky:
      "radial-gradient(1000px 700px at 20% 15%, rgba(210,235,255,.90) 0%, rgba(210,235,255,0) 60%),\
       radial-gradient(800px 600px at 80% 70%, rgba(180,205,230,.35) 0%, rgba(180,205,230,0) 60%),\
       linear-gradient(180deg,#f8fbff 0%, #ffffff 68%)",

  // 暖白纸感 → 对内容最友好，基本不喧宾夺主
  warmPaper:
      "radial-gradient(900px 600px at 25% 20%, rgba(255,244,223,.85) 0%, rgba(255,244,223,0) 55%),\
       linear-gradient(180deg,#fffaf2 0%, #ffffff 72%)",

  // 夜间柔和深蓝（不是纯黑，阅读友好）
  nightSoft:
      "radial-gradient(900px 650px at 30% 20%, rgba(60,85,120,.45) 0%, rgba(60,85,120,0) 55%),\
       linear-gradient(180deg,#1a2430 0%, #202b38 70%)"
};
//加载自定义背景图片前的背景
bgImage.value = bgPresets.softSky;

const bgStyle = computed(() => {
  const v = (bgImage.value || '').trim();
  if (!v) return {}; // 用你 .person-info 里的默认图
  // 如果是 gradient(...) 或多层背景字符串，直接塞进去
  if (/gradient\(/i.test(v) || v.includes(',')) {
    return { backgroundImage: v };
  }
  // 其他情况按 URL 处理
  return { backgroundImage: `url('${v}')` };
});

// 左侧面板
const dockCollapsed = ref(false);

// 音频
const bgAudioRef = ref(null);
const bgAudio = ref("");
const bgAudioInput = ref("");
const isAudioPlaying = ref(false);
const audioVolume = ref(70);
const audioMuted = ref(false);

// ==== 背景图片：文件/URL/重置 ====

const applyBgImageUrl = () => {
  if (!bgImageInput.value) {
    ElMessage.error("无效路径!");
    return false;
  }
  bgImage.value = bgImageInput.value.trim();
  let result = sendAxiosRequest("/blog-api/userInformation/setPersonInfo", {
    userCode: userStore.userBean.code,
    fieldName: "BGIMAGEURL",
    fieldValue: bgImage.value
  });
  ElMessage.success("背景图片已应用");
};

const resetBgImage = () => {
  bgImage.value = "";
  ElMessage.success("已恢复默认背景");
};

const previewBgImage = () => {
  if (!bgImage.value) return;
  window.open(bgImage.value, "_blank");
};

// ==== 背景音乐：文件/URL/控制 ====
const applyBgAudioUrl = () => {
  if (!bgAudioInput.value) {
    ElMessage.error("无效路径!");
    return false;
  }
  let result = sendAxiosRequest("/blog-api/userInformation/setPersonInfo", {
    userCode: userStore.userBean.code,
    fieldName: "BGMUSICURL",
    fieldValue: bgAudioInput.value.trim()
  });
  setAudioSrc(bgAudioInput.value.trim());
  ElMessage.success("背景音乐已设置");
};

function setAudioSrc(src) {
  bgAudio.value = src;
  // 若已在播放，切换后继续播
  nextTickPlayIfWanted();
}

const toggleAudio = async () => {
  const el = bgAudioRef.value;
  if (!el) return;
  if (isAudioPlaying.value) {
    el.pause();
    isAudioPlaying.value = false;
  } else {
    try {
      await el.play();
      isAudioPlaying.value = true;
    } catch (e) {
      ElMessage.error("浏览器阻止自动播放，请手动点击播放");
    }
  }
};

const applyVolume = () => {
  const el = bgAudioRef.value;
  if (!el) return;
  el.volume = Math.min(1, Math.max(0, audioVolume.value / 100));
};

const applyMute = () => {
  const el = bgAudioRef.value;
  if (!el) return;
  el.muted = !!audioMuted.value;
};

const nextTickPlayIfWanted = async () => {
  await Promise.resolve();
  const el = bgAudioRef.value;
  if (!el) return;
  el.volume = Math.min(1, Math.max(0, audioVolume.value / 100));
  el.muted = !!audioMuted.value;
  // 不强制自动播放，尊重用户交互；保留按钮控制
};

async function setBgImageSafely(url) {
  if (!url) { bgImage.value = ''; return; }
  const img = new Image();
  img.src = url;
  try {
    if (img.decode) await img.decode();
    else await new Promise(r => { img.onload = r; img.onerror = r; });
  } finally {
    bgImage.value = url;
    await nextTick();
    const root = document.querySelector('.person-info');
    if (root) {
      root.classList.add('bg-fade');
      setTimeout(() => root.classList.remove('bg-fade'), 260);
    }
  }
}

//获取用户主页信息  背景图片 背景音乐等等
const setPersonInfo = async () => {
  let result = await sendAxiosRequest("/blog-api/userInformation/getPersonInfo", {userCode: user.value.code});
  if (result && !result.isError) {
    result = result.result[0];
    //背景音乐
    bgAudioInput.value = result.BGMUSICURL || "";
    bgAudio.value = result.BGMUSICURL || "";
    if (bgAudio.value) {
      attemptAutoplay();
    }
    //背景图片
    await setBgImageSafely(result.BGIMAGEURL || "");
    bgImageInput.value = result.BGIMAGEURL || "";
    bgImage.value = result.BGIMAGEURL || "";

  }
}


async function attemptAutoplay() {
  const el = bgAudioRef.value;
  if (!el || !bgAudio.value) return;

  // 先同步应用音量/静音
  el.volume = Math.min(1, Math.max(0, audioVolume.value / 100));
  el.muted = !!audioMuted.value;

  try {
    // ① 正常尝试
    await el.play();
    isAudioPlaying.value = true;
  } catch (e1) {
    try {
      // ② 静音自动播（大多数浏览器允许）
      el.muted = true;
      await el.play();
      isAudioPlaying.value = true;

      // 等用户第一次交互后再恢复你设定的静音状态
      const unmuteOnce = () => {
        el.muted = !!audioMuted.value;
        document.removeEventListener('pointerdown', unmuteOnce);
      };
      document.addEventListener('pointerdown', unmuteOnce, {once: true});
    } catch (e2) {
      // ③ 最后兜底：等待一次用户手势再播放
      const resume = () => {
        el.play().then(() => {
          isAudioPlaying.value = true;
          document.removeEventListener('pointerdown', resume);
        }).catch(() => {
        });
      };
      document.addEventListener('pointerdown', resume, {once: true});
    }
  }
}

// 恢复播放状态
onMounted(() => {

  // 初始化音量/静音
  applyVolume();
  applyMute();
});
</script>

<style scoped>

.person-info {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  max-height: 100%;
  padding: 24px 0;
  box-sizing: border-box;
  position: relative;
  /*background-color: #f8f8f8;*/
  background: url('/picture/user/default.webp') no-repeat center center fixed;
  background-size: cover; /* 关键：保持比例并填满容器 */
  background-attachment: fixed; /* 在手机上可能会失效，可根据需求调整 */

  transition: background-image .001s; /* 防止重排，几乎无感 */
}

.person-info.bg-fade {
  animation: bgfade .25s ease;
}
@keyframes bgfade {
  from { opacity: .6; }
  to   { opacity: 1; }
}

.inner-container {
  max-width: 900px;
  margin: 0 auto;
  padding: 0 16px;
  box-sizing: border-box;
}

.toolbar {
  text-align: right;
  margin-bottom: 12px;
}

.section-card,
.profile-card {
  background-color: rgba(255, 255, 255) !important;
  border-radius: 12px;
  box-shadow: 0 4px 18px rgba(0, 0, 0, 0.08);
  opacity: 0.9; /*透明度*/
}

.profile-card {
  margin-bottom: 24px;
  padding: 24px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
}

.profile-header {
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 20px;
}

.left-info {
  display: flex;
  align-items: center;
}

.author-avatar {
  width: 100px !important;
  height: 100px !important;
  font-size: 20px;
  background-color: #f2f2f2;
}

.profile-details {
  margin-left: 24px;
}

.username {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 24px;
  font-weight: bold;
}
.public-badge{
  color: #000;
  font-size: 12px;
  font-weight: bold;
  border-radius: 10px;
  padding: 2px 8px; /* 稍微增加了点左右内边距，更好看 */
  line-height: 1.2;

  /* --- 核心位置微调 --- */
  margin-left: 4px; /* 和名字拉开一点距离 */
  flex-shrink: 0;   /* 关键：保证名字再长，标签都不会被压缩变形 */
  position: relative;
  top: 0px;        /* 往上提一点，制造右上角上标的视觉效果 */
}

.superAdmin-badge {
  background-color: #ffdf02;

}

.admin-badge {
  background-color: #86ff93;
}
.ban-badge {
  background-color: #ff5a5a;
}
.user-remark {
  font-size: 14px;
  color: #555;
  margin: 4px 0;
  max-width: 300px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  background-color: #f8f8f8;
  padding: 4px 8px;
  border-radius: 6px;
  font-style: italic;
}

.userip {
  color: #888;
  margin: 4px 0;
}

.right-info {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-end;
  gap: 10px;
}

.stats {
  font-size: 14px;
  color: #666;
  display: flex;
  gap: 24px;
}

.author-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.author-actions .el-button {
  box-shadow: 0 2px 8px rgb(64 158 255 / 0.3);
  font-weight: 600;
  user-select: none;
}

.follow-button {
  margin-top: 12px;
  transition: all 0.3s ease;
}

.section-card {
  margin-bottom: 30px;
  padding: 16px 20px;
  border-radius: 12px;
  background-color: #ffffff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
}

/* 新增：标题 + 仅文章开关区域 */
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.section-title {
  font-size: 18px;
  font-weight: bold;
  color: #333;
}

.blog-card {
  transition: 0.3s;
  cursor: pointer;
}

.blog-card:hover {
  transform: scale(1.01);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}

.blog-summary {
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.6;
  color: #666;
  margin-top: 8px;
}

.file-table :deep(.el-button) {
  font-weight: bold;
}

.collapse-toggle {
  margin-top: 12px;
  text-align: center;
}

/* 弹窗卡片列表通用 */
.card-list-scroll {
  max-height: 560px;
  overflow-y: auto;
  padding: 12px 16px;
  box-sizing: border-box;
}

.card-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.list-card {
  padding: 12px 16px;
  border-radius: 8px;
  background-color: #edf8f4;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s ease;
}

.list-card:hover {
  transform: scale(1.01);
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.08);
}

.card-top {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #666;
}

.card-body {
  margin-top: 6px;
}

.card-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 6px;
  color: #333;
}

.card-summary {
  font-size: 13px;
  line-height: 1.6;
  color: #555;
  -webkit-line-clamp: 2;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.list-card-follow {
  padding: 12px 16px;
  background-color: #ffffff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}

.follower-card-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.author-avatar-follow {
  width: 60px !important;
  height: 60px !important;
  font-size: 20px;
  background-color: #f2f2f2;
  flex-shrink: 0;
}

.author-name {
  font-size: 14px;
  font-weight: bold;
  color: #333;
}

.loading-text,
.end-text {
  text-align: center;
  color: #999;
  margin: 16px 0;
}

/* 左侧悬浮设置面板 */
.left-dock {
  position: fixed;
  left: 20px;
  top: 120px;
  width: 260px;
  z-index: 20;
  transition: transform 0.25s ease, opacity 0.25s ease;
}

.left-dock.collapsed {
  transform: translateX(-280px);
}

.dock-card {
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: saturate(180%) blur(6px);
  border-radius: 12px;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.12);
  padding-bottom: 12px;
}

.dock-title {
  font-weight: 700;
  color: #333;
  margin-bottom: 8px;
}

.dock-subtitle {
  font-weight: 600;
  font-size: 13px;
  color: #666;
  margin-bottom: 6px;
}

.dock-block {
  margin-bottom: 12px;
}

.dock-upload {
  display: inline-block;
}

.dock-actions {
  display: flex;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.audio-row {
  padding-top: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.vol-label {
  font-size: 12px;
  color: #666;
  width: 36px;
}

/* 折叠按钮 */
.dock-toggle {
  position: absolute;
  right: -18px;
  top: 10px;
  width: 22px;
  height: 28px;
  border-radius: 0 6px 6px 0;
  border: 1px solid #dcdfe6;
  background: #fff;
  color: #666;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* —— 更小更精致的迷你音乐条 —— */
.music-mini {
  position: fixed;
  left: 20px;
  top: 20px;
  width: 220px; /* 更窄 */
  z-index: 19;
  padding: 8px 10px; /* 更小的内边距 */
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: saturate(180%) blur(6px);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.12);
  transition: transform .18s ease, box-shadow .18s ease, background .18s ease;
}

.music-mini__title {
  font-size: 12px; /* 更小字号 */
  font-weight: 600;
  color: #333;
  margin-bottom: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.music-mini__row {
  display: flex;
  align-items: center;
  gap: 8px; /* 更紧凑 */
}

/* 圆形图标按钮（更小巧） */
.icon-btn {
  width: 28px;
  height: 28px;
  padding: 0;
  font-size: 14px;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

/* 精简滑条：更细更小的滑块 */
.mini-slider {
  flex: 1;
  margin-left: 6px;
}

.mini-slider :deep(.el-slider__runway) {
  height: 3px; /* 细一些 */
  border-radius: 999px;
}

.mini-slider :deep(.el-slider__bar) {
  height: 3px;
  border-radius: 999px;
}

.mini-slider :deep(.el-slider__button-wrapper) {
  width: 12px;
  height: 12px;
}

.mini-slider :deep(.el-slider__button) {
  width: 10px;
  height: 10px; /* 小滑块 */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.15);
}

@media (max-width: 768px) {
  /* 时间线整体去掉左边距和内边距 */
  .person-info :deep(.el-timeline) {
    padding-left: 0 !important;
    margin-left: 0 !important;
  }

  /* 时间线每一项去除左内边距和左边距 */
  .person-info :deep(.el-timeline-item) {
    padding-left: 0 !important;
    margin-left: 0 !important;
  }

  /* 去除内容wrapper左边距，宽度撑满 */
  .person-info :deep(.el-timeline-item__wrapper) {
    margin-left: 0 !important;
    width: 100% !important;
    padding-left: 0 !important;
  }

  /* 隐藏竖线和节点 */
  .person-info :deep(.el-timeline-item__tail),
  .person-info :deep(.el-timeline-item__node) {
    display: none !important;
  }

  /* 个人信息头部布局调整 */
  .profile-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .left-info {
    flex-direction: column;
    align-items: center;
  }

  .profile-details {
    margin-left: 0;
    margin-top: 12px;
  }

  .right-info {
    align-items: center;
    justify-content: center;
    margin-top: 12px;
  }

  .stats {
    justify-content: center;
  }

  .interaction-buttons {
    flex-direction: column;
    align-items: center;
  }

  .follow-button {
    width: 100%;
  }

  .section-title {
    text-align: center;
  }

  .list-card, .list-card-follow {
    height: auto;
    flex-direction: column;
    align-items: flex-start;
  }

  .card-top {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .card-body {
    width: 100%;
  }

  .card-summary {
    -webkit-line-clamp: 4;
  }

  .left-dock {
    transform: translateX(-280px);
  }

  .left-dock.collapsed {
    transform: translateX(-280px); /* 保持折叠 */
  }

  .left-dock .dock-toggle {
    transform: translateX(0px);
  }

  .music-mini {
    left: 50%;
    transform: translateX(-50%);
    top: 12px;
    width: calc(100vw - 100px);
    padding: 8px 10px;
  }
}
</style>

