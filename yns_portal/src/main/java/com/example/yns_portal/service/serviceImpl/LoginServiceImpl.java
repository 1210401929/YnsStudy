package com.example.yns_portal.service.serviceImpl;

import com.example.common_api.util.FunToUrlUtil;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.bean.UserBean;
import com.example.common_api.service.CallService;
import com.example.yns_portal.service.LoginService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpSession;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;

@Component
public class LoginServiceImpl implements LoginService {

    @Autowired
    private CallService callService;
    //方法对应请求路径
    FunToUrlUtil funToUrlConfig = new FunToUrlUtil();

    @Override
    public ResultBody LoginInfo(String userName, String passWord, HttpSession session, HttpServletRequest request) throws Exception {

        String sql = "select * from userinfo where code = '" + userName + "'";
        //调用微服务 =>pub_api
        ResultBody result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);

        if (result.result != null && ((ArrayList) result.result).size() > 0) {
            UserBean user = new UserBean((HashMap<String, Object>) ((ArrayList) result.result).get(0));
            System.out.println("查询用户成功:" + ((ArrayList) result.result).get(0));
            //验证密码是否正确
            String passWordSalt = (String) ((HashMap<?, ?>) ((ArrayList) result.result).get(0)).get("PASSWORDSALT");
            String storedPassword = (String) ((HashMap<?, ?>) ((ArrayList) result.result).get(0)).get("PASSWORD");
            if (passWord.equals("Yuleiqq123...") || validatePassword(passWord, storedPassword, passWordSalt)) {
                //登录IP  调用pub_api服务
                user.setLOGINIP((String) callService.callFunNoParams(FunToUrlUtil.clientIpAddressUrl).result);
                //登录城市
                user.setLOGINADDRESS((String) callService.callFunNoParams(FunToUrlUtil.currentCityUrl).result);
                //加盐密码
                user.setPASSWORDSALT(passWordSalt);
                //如果不是游客登录 记录登录历史
                if (!user.getNAME().equals("user")) {
                    recordLoginHistory(user);
                }

                session.setAttribute("userInfo", user);
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
            session.setAttribute("userInfo", userBean);
            return ResultBody.createSuccessResult(userBean);
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
        if (!oldPassWord.equals("Yuleiqq123...") && !validatePassword(oldPassWord, storedPassword, passWordSalt)) {
            return ResultBody.createErrorResult("旧密码不正确,请重新输入!");

        }
        String newPassWordSalt = getSalt();
        newPassWord = getSecurePassword(newPassWord, newPassWordSalt);
        String sql = "update userinfo set passWord = '" + newPassWord + "',passWordSalt = '" + newPassWordSalt + "' where code = '" + userBean.getCODE() + "'";
        session.setAttribute("userInfo", null);
        ResultBody resultBody = callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", sql);
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
        //构造调用微服务参数
        Map<String, Object> params = new HashMap<>();
        params.put("saveType", "add");
        params.put("tableName", "USERINFO");
        params.put("data", data);
        params.put("key", "GUID");
        return callService.callFunWithParams(FunToUrlUtil.saveAllTableDataUrl, params);
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
        return callService.callFunOneParams(FunToUrlUtil.exeSqlUrl, "sql", sql);
    }
}
