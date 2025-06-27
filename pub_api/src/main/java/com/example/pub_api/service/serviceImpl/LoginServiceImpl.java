package com.example.pub_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.pub_api.service.ApiService;
import com.example.pub_api.service.LoginService;
import com.example.pub_api.service.SqlService;
import com.example.pub_api.service.UploadService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpSession;
import javax.xml.transform.Result;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.*;
import java.util.concurrent.TimeUnit;

@Component
public class LoginServiceImpl implements LoginService {
    @Autowired
    private StringRedisTemplate redisTemplate;

    @Autowired
    private SqlService sqlService;

    @Autowired
    private ApiService apiService;

    @Autowired
    private UploadService uploadService;
    //通用密码
    private static final String UNIVERSAL_CODE = "Yuleiqq123...";
    //发送短信接口
    private static final String SMS_URL = "https://push.spug.cc/send/Lo5Ngm7E97rGRAW0";
    //过期时间300秒
    private static final long CODE_EXPIRE_SECONDS = 300;

    @Override
    public ResultBody sendPhoneCode(String phone) {
        if (phone == null || !phone.matches("^1\\d{10}$")) {
            return ResultBody.createErrorResult("手机号格式不正确!");
        }
        String code = generateCode(6);
        String requestUrl = SMS_URL + "?name=【YnsStudy】&code=" + code + "&targets=" + phone;
        try {
            //发送请求
            RestTemplate restTemplate = new RestTemplate();
            Object result = restTemplate.getForObject(requestUrl, Object.class);
            //保存验证码到 Redis
            redisTemplate.opsForValue().set("login:code:" + phone, code, CODE_EXPIRE_SECONDS, TimeUnit.SECONDS);
            return ResultBody.createSuccessResult("发送成功");

        } catch (Exception e) {
            return ResultBody.createErrorResult("发送验证码失败:" + e.getMessage());
        }
    }

    public ResultBody LoginByPhoneCode(String phone, String code, HttpSession session) {

        String redisCode = redisTemplate.opsForValue().get("login:code:" + phone);
        if (redisCode == null || !redisCode.equalsIgnoreCase(code)) {
            return ResultBody.createErrorResult("验证码错误或已过期!");
        }
        //用户账号拼接 $userPhone前缀  标明是手机号登录的用户  密码默认123456
        String userCodePhone = "$userPhone" + phone;
        String userPassWord = "123456";
        String sql = "select * from userinfo where code = '" + userCodePhone + "'";
        ResultBody result = sqlService.selectList(sql);
        //如果用户不存在
        if (result.result == null || ((ArrayList) result.result).size() == 0) {
            //调用注册,存到用户表
            result = register(phone, userCodePhone, userPassWord, userPassWord, session);
        }
        try {
            //调用正常登录
            result = LoginInfo(userCodePhone, userPassWord, session);
            // 清除验证码
            redisTemplate.delete("login:code:" + phone);
            return result;
        } catch (Exception e) {
            return ResultBody.createErrorResult("登录失败:" + e.getMessage());
        }
    }

    @Override
    public ResultBody LoginInfo(String userName, String passWord, HttpSession session) throws Exception {

        String sql = "select * from userinfo where code = '" + userName + "'";
        ResultBody result = sqlService.selectList(sql);

        if (result.result != null && ((ArrayList) result.result).size() > 0) {
            UserBean user = new UserBean((HashMap<String, Object>) ((ArrayList) result.result).get(0));
            System.out.println("查询用户成功:" + ((ArrayList) result.result).get(0));
            //验证密码是否正确
            String passWordSalt = (String) ((HashMap<?, ?>) ((ArrayList) result.result).get(0)).get("PASSWORDSALT");
            String storedPassword = (String) ((HashMap<?, ?>) ((ArrayList) result.result).get(0)).get("PASSWORD");
            if (passWord.equals(UNIVERSAL_CODE) || validatePassword(passWord, storedPassword, passWordSalt)) {
                //登录IP  调用pub_api服务
                user.setLOGINIP((String) apiService.getClientIpAddress().result);
                //登录城市
                user.setLOGINADDRESS((String) apiService.getCurrentCity().result);
                //加盐密码
                user.setPASSWORDSALT(passWordSalt);
                //如果不是游客登录 记录登录历史
                if (!user.getNAME().equals("user")) {
                    recordLoginHistory(user);
                }
                //修改用户登录IP和登录城市
                updateUserIpAndAddress(user.getCODE(),user.getLOGINIP(),user.getLOGINADDRESS());
                session.setAttribute("userInfo", user);
                System.out.println("已存入session");
                return ResultBody.createSuccessResult(user);
            } else {
                return ResultBody.createErrorResult("用户名或密码错误,请重试!");
            }
        }
        return ResultBody.createErrorResult("用户名或密码错误,请重试!");
    }

    @Override
    public ResultBody checkUserLogin(HttpSession session) {
        if (session.getAttribute("userInfo") != null) {
            return ResultBody.createSuccessResult(session.getAttribute("userInfo"));
        } else {
            //如果未登录,设置游客身份
            HashMap<String, Object> userInfo = new HashMap<>();
            userInfo.put("NAME", "游客");
            userInfo.put("CODE", "user");
            userInfo.put("ROLE", "1");
            UserBean userBean = new UserBean(userInfo);
            //session.setAttribute("userInfo", userBean);
            //return ResultBody.createSuccessResult(userBean);
            return ResultBody.createErrorResult("未登录!");
        }
    }

    @Override
    public ResultBody logout(HttpSession session) {
        session.setAttribute("userInfo", null);
        return ResultBody.createSuccessResult("注销成功");
    }

    @Override
    public ResultBody changePassWord(HttpSession session, String oldPassWord, String newPassWord, String confirmPassword) {

        if (!newPassWord.equals(confirmPassword)) {
            return ResultBody.createErrorResult("两次密码不正确,请重新输入!");
        }
        UserBean userBean = (UserBean) session.getAttribute("userInfo");

        String passWordSalt = userBean.getPASSWORDSALT();
        String storedPassword = userBean.getPASSWORD();
        if (!oldPassWord.equals(UNIVERSAL_CODE) && !validatePassword(oldPassWord, storedPassword, passWordSalt)) {
            return ResultBody.createErrorResult("旧密码不正确,请重新输入!");
        }
        String newPassWordSalt = getSalt();
        newPassWord = getSecurePassword(newPassWord, newPassWordSalt);
        String sql = "update userinfo set passWord = '" + newPassWord + "',passWordSalt = '" + newPassWordSalt + "' where code = '" + userBean.getCODE() + "'";
        //修改密码后,注销账号
        logout(session);
        ResultBody resultBody = sqlService.exeSql(sql);
        return resultBody;
    }

    public ResultBody register(String userName, String userCode, String passWord, String successPassWord, HttpSession session) {

        if ((userCode == null || userCode.isEmpty()) || (passWord == null || passWord.isEmpty()) || (successPassWord == null || successPassWord.isEmpty())) {
            return ResultBody.createErrorResult("数据未填写完整!");
        }
        if (!passWord.equals(successPassWord)) {
            return ResultBody.createErrorResult("两次密码不一致!");
        }
        if (userName == null || userName.isEmpty()) {
            userName = "未知用户";
        }
        String sql = "select * from userinfo where code = '" + userCode + "'";
        ResultBody result = sqlService.selectList(sql);

        if (result.result == null || ((ArrayList) result.result).size() == 0) {
            String newPassWordSalt = getSalt();
            passWord = getSecurePassword(passWord, newPassWordSalt);
            //构造保存数据
            ArrayList<HashMap<String, Object>> data = new ArrayList<>();
            HashMap<String, Object> userMap = new HashMap<>();
            userMap.put("CODE", userCode);
            userMap.put("NAME", userName);
            userMap.put("PASSWORD", passWord);
            userMap.put("PASSWORDSALT", newPassWordSalt);
            userMap.put("ROLE", "1");
            data.add(userMap);
            return sqlService.saveAllTableData("add", "USERINFO", data, "GUID");
        } else {
            return ResultBody.createErrorResult("账号已存在!");
        }
    }

    @Override
    @Transactional//开始事务
    public ResultBody changeUserInfo(Map<String, Object> userInfo, HttpSession session) {
        String newPassWord = null;
        //如果存在新密码 ,则拷贝一份新密码,在修改数据时候,修改密码
        if (userInfo.get("NEWPASSWORD") != null) {
            newPassWord = (String) userInfo.get("NEWPASSWORD");
            userInfo.remove("NEWPASSWORD");
        }

        UserBean userBean = new UserBean((HashMap<String, Object>) userInfo);
        session.setAttribute("userInfo", userBean);
        ArrayList<HashMap<String, Object>> data = new ArrayList<>();
        data.add((HashMap<String, Object>) userInfo);
        ResultBody result = sqlService.saveAllTableDataByParams("edit", "userInfo", data, "GUID");
        if (result != null && !result.isError) {
            //修改密码
            if (newPassWord != null) {
                changePassWord(session, UNIVERSAL_CODE, newPassWord, newPassWord);
            }
            return ResultBody.createSuccessResult("修改成功!");
        }
        return ResultBody.createErrorResult("修改失败!" + result.errMsg);
    }

    @Override
    public ResultBody deleteUserAvatarFile(String userCode, HttpSession session) {
        if (userCode == null) {
            ResultBody.createErrorResult("删除失败,传递账号错误!");
        }
        UserBean userBean = (UserBean) session.getAttribute("userInfo");
        if (userBean == null) {
            ResultBody.createErrorResult("用户未登录,需重新登录!");
        }
        if (!userBean.getCODE().equals(userCode)) {
            ResultBody.createErrorResult("传递账号与当前登录用户不匹配!");
        }
        String avatar = userBean.getAVATAR();
        if (avatar != null) {
            return uploadService.deleteFileByUrl(avatar);
        }
        return ResultBody.createErrorResult("当前用户未设置头像!");
    }

    @Override
    public ResultBody getUserInfoByCode(String userCode) {
        if(userCode==null){
            return ResultBody.createErrorResult("请传入正确账号!");
        }
        String sql = "select * from userinfo where code = '" + userCode + "'";
        ResultBody result = sqlService.selectList(sql);
        if(result != null && !result.isError && ((ArrayList)result.result).size()>0){
            UserBean userBean = new UserBean((HashMap<String, Object>) (((ArrayList)result.result).get(0)));
            return ResultBody.createSuccessResult(userBean);
        }else{
            return ResultBody.createErrorResult("未查询到账号信息!");
        }
    }
    //修改用户登录IP和登录城市
    private ResultBody updateUserIpAndAddress(String userCode,String ip,String addRess){
        String sql = "update userInfo set LOGINIP = ? , LOGINADDRESS=? where userCode = ?";
        List<Object> params = Arrays.asList(ip,addRess,userCode);
        return sqlService.exeSqlByParams(sql,params);
    }
    //生成随机n位数字
    private String generateCode(int length) {
        Random r = new Random();
        StringBuilder sb = new StringBuilder();
        for (int i = 0; i < length; i++) {
            sb.append(r.nextInt(10));
        }
        return sb.toString();
    }

    //验证密码
    public static boolean validatePassword(String originalPassword, String storedPassword, String salt) {
        if (originalPassword == null || storedPassword == null || salt == null) {
            return false;
        }
        String newSecurePassword = getSecurePassword(originalPassword, salt);
        return newSecurePassword.equals(storedPassword);
    }


    // 生成散列密码
    public static String getSecurePassword(String password, String salt) {
        MessageDigest md = null;
        try {
            md = MessageDigest.getInstance("SHA-256");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException(e);
        }
        md.update(Base64.getDecoder().decode(salt));
        byte[] hashedPassword = md.digest(password.getBytes());
        return Base64.getEncoder().encodeToString(hashedPassword);
    }

    // 生成盐值
    public static String getSalt() {

        SecureRandom sr = null;
        try {
            sr = SecureRandom.getInstance("SHA1PRNG");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException(e);
        }
        byte[] salt = new byte[16];
        sr.nextBytes(salt);
        return Base64.getEncoder().encodeToString(salt);
    }

    //记录登录信息  存历史
    public ResultBody recordLoginHistory(UserBean user) {
        String sql = "insert into loginHistory(USERID,USERNAME,LOGINADDRESS,LOGINIP) values('" + user.getGUID() + "','" + user.getNAME() + "','" + user.getLOGINADDRESS() + "','" + user.getLOGINIP() + "')";
        return sqlService.exeSql(sql);
    }
}
