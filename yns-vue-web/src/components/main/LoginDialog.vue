<template>
  <div>
    <!-- 登录按钮区域 -->
    <div>
      <template v-if="userStore.userBean.name">
        <el-dropdown @command="handleCommand">
    <span class="el-dropdown-link login-button" style="display: flex; align-items: center; gap: 8px;">
                    <el-avatar
                        :src="userStore.userBean.avatar"
                        size="large"
                        class="dropdown-avatar"
                        alt="用户头像"
                    >
                {{ userStore.userBean.name?.charAt(0) }}
              </el-avatar>
      <el-icon><User/></el-icon>
      {{ userStore.userBean.name }}
    </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人中心</el-dropdown-item>
              <el-dropdown-item command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>

      <template v-else>
        <div class="login-button" @click="visible = true">
          <el-icon>
            <User/>
          </el-icon>
          登录
        </div>
      </template>
    </div>

    <!-- 用户名密码登录框 -->
    <teleport to="body">
      <el-dialog
          v-model="visible"
          width="400px"
          :show-close="false"
          :close-on-click-modal="false"
          class="custom-dialog"
          center
      >
        <template #header>
          <span style="font-size: 20px;">用户登录</span>
        </template>

        <el-form :model="form" ref="formRef" :rules="rules" label-position="top">
          <el-form-item prop="username">
            <el-input v-model="form.username" placeholder="用户名"/>
          </el-form-item>

          <el-form-item prop="password">
            <el-input v-model="form.password" type="password" placeholder="密码" show-password/>
          </el-form-item>

          <el-form-item>
            <el-button type="success" @click="onLogin" style="width: 100%;">登录</el-button>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" @click="onRegister" style="width: 100%;">注册</el-button>
          </el-form-item>

          <el-form-item>
            <el-button type="danger" @click="visible = false" style="width: 100%;">取消</el-button>
          </el-form-item>

          <div style="margin-top: 14px; text-align: center;">
            <a @click="callPhoneLogin" style="cursor: pointer; color: #4caf50;">
              📱 手机号登录
            </a>
          </div>
        </el-form>
      </el-dialog>
    </teleport>

    <!-- 手机号登录框 -->
    <teleport to="body">
      <el-dialog
          v-model="phoneLoginVisible"
          width="400px"
          :show-close="false"
          :close-on-click-modal="false"
          class="custom-dialog"
          center
      >
        <template #header>
          <span style="font-size: 20px;">手机号登录</span>
        </template>

        <el-form :model="phoneForm" ref="phoneFormRef" :rules="phoneRules" label-position="top">
          <el-form-item prop="phone">
            <el-input v-model="phoneForm.phone" placeholder="请输入手机号"/>
          </el-form-item>

          <el-form-item prop="code">
            <div style="display: flex; gap: 8px; width: 100%;">
              <el-input v-model="phoneForm.code" placeholder="请输入验证码" style="flex: 1;"/>
              <el-button
                  type="primary"
                  @click="sendCode"
                  :disabled="countdown > 0"
                  style="white-space: nowrap;"
              >
                {{ countdown > 0 ? countdown + ' 秒' : '发送验证码' }}
              </el-button>
            </div>
          </el-form-item>

          <el-form-item>
            <el-button type="success" style="width: 100%;" @click="submitPhoneLogin">登录</el-button>
          </el-form-item>

          <el-form-item>
            <el-button type="danger" style="width: 100%;" @click="phoneLoginVisible = false">取消</el-button>
          </el-form-item>

          <div style="margin-top: 10px; text-align: center;">
            <a @click="returnToLogin" style="cursor: pointer; color: #409EFF;">← 返回</a>
          </div>
        </el-form>
      </el-dialog>
    </teleport>
    <!-- 注册框 -->
    <teleport to="body">
      <el-dialog
          v-model="registerVisible"
          width="400px"
          :show-close="false"
          :close-on-click-modal="false"
          class="custom-dialog"
          center
      >
        <template #header>
          <span style="font-size: 20px;">用户注册</span>
        </template>

        <el-form :model="registerForm" ref="registerFormRef" :rules="registerRules" label-position="top">
          <el-form-item prop="name">
            <el-input v-model="registerForm.name" placeholder="请输入用户名"/>
          </el-form-item>

          <el-form-item prop="username">
            <el-input v-model="registerForm.username" placeholder="请输入账号"/>
          </el-form-item>

          <el-form-item prop="password">
            <el-input v-model="registerForm.password" type="password" placeholder="请输入密码"/>
          </el-form-item>

          <el-form-item prop="confirmPassword">
            <el-input v-model="registerForm.confirmPassword" type="password" placeholder="确认密码"/>
          </el-form-item>

          <el-form-item>
            <el-button type="primary" style="width: 100%;" @click="submitRegister">注册</el-button>
          </el-form-item>

          <el-form-item>
            <el-button type="danger" style="width: 100%;" @click="registerVisible = false">取消</el-button>
          </el-form-item>
        </el-form>
      </el-dialog>
    </teleport>
  </div>
</template>

<script setup>
import {ref} from 'vue';
import {useRouter} from "vue-router";
import {ElMessage} from 'element-plus';
import {sendAxiosRequest} from '@/utils/common.js';
import {useUserStore} from '@/stores/main/user.js';
import {useBlogContentStore} from "@/stores/detail/blog.js";
import {User} from '@element-plus/icons-vue';

const router = useRouter();

const visible = ref(false);
const phoneLoginVisible = ref(false);
const countdown = ref(0);
let timer = null;

const blogContentStore = useBlogContentStore();
const userStore = useUserStore();
userStore.initFromLocal();

// 原始账号登录
const formRef = ref();
const form = ref({username: '', password: ''});
const rules = {
  username: [{required: true, message: '请输入用户名', trigger: 'blur'}],
  password: [{required: true, message: '请输入密码', trigger: 'blur'}]
};

const onLogin = async () => {
  try {
    //执行表单验证
    const valid = await new Promise((resolve, reject) => {
      formRef.value.validate(valid => (valid ? resolve(true) : reject(false)));
    });
    if (valid) {
      const {username, password} = form.value;
      const result = await sendAxiosRequest('/pub-api/login/loginUser', {
        userName: username,
        passWord: password
      });
      if (result && result.result && !result.isError) {
        userStore.setUser(result.result);
        ElMessage.success('登录成功！');
        //刷新页面
        location.reload();
        visible.value = false;

      } else {
        ElMessage.error(result.errMsg);
      }
    }
  } catch {
    ElMessage.error('表单验证未通过');
  }
};


// 手机号登录
const phoneFormRef = ref();
const phoneForm = ref({phone: '', code: ''});

const phoneRules = {
  phone: [{required: true, message: '请输入手机号', trigger: 'blur'}],
  code: [{required: true, message: '请输入验证码', trigger: 'blur'}]
};

const callPhoneLogin = () => {
  visible.value = false;
  phoneLoginVisible.value = true;
  phoneForm.value.phone = '';
  phoneForm.value.code = '';
};

const returnToLogin = () => {
  phoneLoginVisible.value = false;
  visible.value = true;
};

// 发送短信验证码
const sendCode = async () => {
  const phone = phoneForm.value.phone;
  if (!/^1\d{10}$/.test(phone)) {
    ElMessage.warning('请输入正确的11位手机号');
    return;
  }

  try {
    let result = await sendAxiosRequest('/pub-api/login/sendPhoneCode', {phone});
    ElMessage.success('验证码已发送');
    countdown.value = 60;
    timer = setInterval(() => {
      countdown.value--;
      if (countdown.value <= 0) clearInterval(timer);
    }, 1000);
  } catch (e) {
    ElMessage.error('验证码发送失败');
  }
};

// 提交手机号验证码登录
const submitPhoneLogin = async () => {
  try {
    const valid = await phoneFormRef.value.validate();
    const res = await sendAxiosRequest('/pub-api/login/loginByPhoneCode', {
      phone: phoneForm.value.phone,
      code: phoneForm.value.code
    });
    if (res && res.result && !res.isError) {
      userStore.setUser(res.result);

      ElMessage.success('登录成功！');
      //刷新页面
      location.reload();
      phoneLoginVisible.value = false;
    } else {
      ElMessage.error(res.errMsg || '登录失败');
    }
  } catch (e) {
    ElMessage.error('请填写手机号和验证码');
  }
};

const handleCommand = async (command) => {
  //个人中心
  if (command === 'profile') {
    const routeUrl = router.resolve({name: 'personalCenter', query: {id: userStore.userBean.code}}).href;
    window.open(routeUrl, userStore.userBean.code);
    //退出登录
  } else if (command === 'logout') {
    await userStore.clearUser();
    //跳转路由到欢迎页面  这里使用replace而不适用push的原因是,实现用户不能通过浏览器进行回退
    await router.replace({name: "Welcome"});
    ElMessage.success('已退出登录');
  }
};
//--------------------------注册相关---------------------------------------
// 注册弹窗
const registerVisible = ref(false);
const registerFormRef = ref();
const registerForm = ref({
  name: '',
  username: '',
  password: '',
  confirmPassword: ''
});

const registerRules = {
  name: [{required: true, message: '请输入用户名', trigger: 'blur'}],
  username: [{required: true, message: '请输入账号', trigger: 'blur'}],
  password: [{required: true, message: '请输入密码', trigger: 'blur'}],
  confirmPassword: [
    {required: true, message: '请确认密码', trigger: 'blur'},
    {
      validator: (rule, value, callback) => {
        if (value !== registerForm.value.password) {
          callback(new Error('两次输入的密码不一致'));
        } else {
          callback();
        }
      },
      trigger: 'blur'
    }
  ]
};

const onRegister = () => {
  visible.value = false;
  registerVisible.value = true;
  registerForm.value = {
    name: '',
    username: '',
    password: '',
    confirmPassword: ''
  };
};

const submitRegister = async () => {
  try {
    //表单验证
    await registerFormRef.value.validate();

    const {name, username, password, confirmPassword} = registerForm.value;

    const result = await sendAxiosRequest('/pub-api/login/register', {
      userName: name,
      userCode: username,
      passWord: password,
      successPassWord: confirmPassword
    });

    if (result && result.result && !result.isError) {
      ElMessage.success('注册成功，请登录');
      registerVisible.value = false;
      visible.value = true;
    } else {
      ElMessage.error(result.errMsg || '注册失败');
    }
  } catch (e) {
    ElMessage.error('请完整填写注册信息');
  }
};

</script>

<style scoped>
.login-button {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 16px;
}

.custom-dialog .el-dialog {
  top: 50% !important;
  transform: translateY(-50%);
}

.dropdown-avatar {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  object-fit: cover;
}

</style>
