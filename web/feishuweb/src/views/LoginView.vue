<template>
  <div class="login-register">
    <h1 class="title">{{ isLoginMode ? '登录' : '注册' }}</h1>
    <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="手机号" prop="phone">
        <el-input v-model="form.phone" placeholder="请输入手机号" />
      </el-form-item>

      <!-- 注册时显示验证码 -->
      <el-form-item label="图形验证码" v-if="!isLoginMode" prop="captchaCode">
        <div style="display: flex; align-items: center;">
          <el-input
              v-model="form.captchaCode"
              placeholder="请输入验证码"
              style="flex: 1; margin-right: 10px"
          />
          <img
              :src="captchaImage"
              @click="refreshCaptcha"
              alt="验证码"
              style="cursor: pointer; height: 40px"
          />
        </div>
      </el-form-item>

      <el-form-item label="密码" prop="password">
        <el-input v-model="form.password" type="password" placeholder="请输入密码" />
      </el-form-item>

      <el-form-item label="确认密码" v-if="!isLoginMode" prop="confirmPassword">
        <el-input v-model="form.confirmPassword" type="password" placeholder="请确认密码" />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" :loading="submitLoading" @click="submitForm" style="width: 100%">
          {{ isLoginMode ? '登录' : '注册' }}
        </el-button>
      </el-form-item>

      <el-form-item>
        <el-link @click="toggleMode" style="float: right">
          {{ isLoginMode ? '没有账号？去注册' : '已有账号？去登录' }}
        </el-link>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import request from '@/utils/request';

export default {
  name: 'LoginRegisterView',
  data() {
    return {
      isLoginMode: true,
      submitLoading: false,
      captchaImage: '',
      form: {
        phone: '',
        password: '',
        confirmPassword: '',
        captchaCode: '',
        captchaId: ''
      },
      rules: {
        phone: [
          { required: true, message: '请输入手机号', trigger: 'blur' },
          { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 6, max: 20, message: '密码长度为6~20位', trigger: 'blur' }
        ],
        captchaCode: [{ required: true, message: '请输入图形验证码', trigger: 'blur' }],
        confirmPassword: [
          { required: true, message: '请确认密码', trigger: 'blur' },
          {
            validator: (rule, value, callback) => {
              if (value !== this.form.password) {
                callback(new Error('两次输入密码不一致'));
              } else {
                callback();
              }
            },
            trigger: 'blur'
          }
        ]
      }
    };
  },
  created() {
    if (!this.isLoginMode) this.refreshCaptcha();
  },
  methods: {
    toggleMode() {
      this.isLoginMode = !this.isLoginMode;
      this.form = {
        phone: '',
        password: '',
        confirmPassword: '',
        captchaCode: '',
        captchaId: ''
      };
      this.$nextTick(() => {
        this.$refs.formRef.clearValidate();
      });
      if (!this.isLoginMode) this.refreshCaptcha();
    },
    async refreshCaptcha() {
      try {
        const res = await request.get('/api/captcha/get');
        this.form.captchaId = res.id;
        this.captchaImage = res.base64;
      } catch (e) {
        this.$message.error('获取验证码失败');
      }
    },
    async submitForm() {
      this.$refs.formRef.validate(async (valid) => {
        if (!valid) return;
        this.submitLoading = true;
        try {
          if (this.isLoginMode) {
            const res = await request.post('/api/login', {
              phone: this.form.phone,
              password: this.form.password,
            });
            if (res.code === 0) {
              localStorage.setItem('token', res.data.token);
              localStorage.setItem('userInfo', JSON.stringify(res.data.user));
              localStorage.setItem('tokenTimestamp', Date.now().toString())
              this.$store.commit('SET_TOKEN', res.data.token);
              this.$store.commit('SET_USER_INFO', res.data.user);
              this.$message.success('登录成功');
              this.$router.push('/');
            } else {
              throw new Error(res.message || '登录失败');
            }
          } else {
            const res = await request.post('/api/register', {
              phone: this.form.phone,
              password: this.form.password,
              captcha_id: this.form.captchaId,
              captcha_code: this.form.captchaCode,
            });
            if (res.code === 0) {
              this.$message.success('注册成功，请登录');
              this.toggleMode();
            } else {
              throw new Error(res.message || '注册失败');
            }
          }
        } catch (err) {
          this.$message.error(err.message);
        } finally {
          this.submitLoading = false;
        }
      });
    }
  }
};
</script>

<style scoped>
.login-register {
  max-width: 400px;
  margin: 60px auto;
  padding: 30px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
}
.title {
  text-align: center;
  font-size: 24px;
  color: #2c3e50;
  margin-bottom: 20px;
}
</style>
