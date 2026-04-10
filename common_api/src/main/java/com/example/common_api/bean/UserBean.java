package com.example.common_api.bean;



import com.fasterxml.jackson.annotation.JsonInclude;

import java.io.Serializable;
import java.util.HashMap;

@JsonInclude(JsonInclude.Include.NON_NULL) // 关键：只有不为 null 的字段才包含在 JSON 中
public class UserBean implements Serializable {

    private String GUID;
    private String CODE;
    private String NAME;
    private String PASSWORD;
    private String PASSWORDSALT;


    private String ROLE;
    private String REMARK;
    private String LOGINADDRESS;
    private String LOGINIP;

    //手机号
    private String PHONE;
    //邮箱
    private String EMAIL;
    //头像
    private String AVATAR;
    //是否禁用
    private String ISBAN;
    //用户唯一数字
    private int USERNUM;

    public String getPHONE() {
        return PHONE;
    }

    public void setPHONE(String PHONE) {
        this.PHONE = PHONE;
    }

    public String getEMAIL() {
        return EMAIL;
    }

    public void setEMAIL(String EMAIL) {
        this.EMAIL = EMAIL;
    }

    public String getAVATAR() {
        return AVATAR;
    }

    public void setAVATAR(String AVATAR) {
        this.AVATAR = AVATAR;
    }

    public String getLOGINADDRESS() {
        return LOGINADDRESS;
    }
    public void setLOGINADDRESS(String LOGINADDRESS) {
        this.LOGINADDRESS = LOGINADDRESS;
    }
    public String getLOGINIP() {
        return LOGINIP;
    }
    public String getPASSWORDSALT() {
        return PASSWORDSALT;
    }

    public void setPASSWORDSALT(String PASSWORDSALT) {
        this.PASSWORDSALT = PASSWORDSALT;
    }

    public String getGUID() {
        return GUID;
    }

    public void setGUID(String GUID) {
        this.GUID = GUID;
    }

    public String getCODE() {
        return CODE;
    }

    public void setCODE(String CODE) {
        this.CODE = CODE;
    }

    public String getNAME() {
        return NAME;
    }

    public void setNAME(String NAME) {
        this.NAME = NAME;
    }

    public String getPASSWORD() {
        return PASSWORD;
    }

    public void setPASSWORD(String PASSWORD) {
        this.PASSWORD = PASSWORD;
    }

    public String getROLE() {
        return ROLE;
    }

    public void setROLE(String ROLE) {
        this.ROLE = ROLE;
    }

    public String getREMARK() {
        return REMARK;
    }

    public void setREMARK(String REMARK) {
        this.REMARK = REMARK;
    }


    public void setLOGINIP(String LOGINIP) {
        this.LOGINIP = LOGINIP;
    }


    public String getISBAN() {
        return ISBAN;
    }

    public void setISBAN(String ISBAN) {
        this.ISBAN = ISBAN;
    }

    public int getUSERNUM() {
        return USERNUM;
    }

    public void setUSERNUM(int USERNUM) {
        this.USERNUM = USERNUM;
    }

    @Override
    public String toString() {
        return "UserBean{" +
                "GUID='" + GUID + '\'' +
                ", CODE='" + CODE + '\'' +
                ", NAME='" + NAME + '\'' +
                ", PASSWORD='" + PASSWORD + '\'' +
                ", PASSWORDSALT='" + PASSWORDSALT + '\'' +
                ", ROLE='" + ROLE + '\'' +
                ", REMARK='" + REMARK + '\'' +
                ", LOGINADDRESS='" + LOGINADDRESS + '\'' +
                ", LOGINIP='" + LOGINIP + '\'' +
                ", PHONE='" + PHONE + '\'' +
                ", EMAIL='" + EMAIL + '\'' +
                ", AVATAR='" + AVATAR + '\'' +
                '}';
    }

    public UserBean(UserBean userinfo) {
        this.GUID =  userinfo.getGUID();
        this.CODE =  userinfo.getCODE();
        this.NAME =  userinfo.getNAME();
        this.PASSWORD =  userinfo.getPASSWORD();
        this.PASSWORDSALT =  userinfo.getPASSWORDSALT();
        this.ROLE =  userinfo.getROLE();
        this.REMARK = userinfo.getREMARK();
        this.PHONE=userinfo.getPHONE();
        this.EMAIL= userinfo.getEMAIL();
        this.AVATAR= userinfo.getAVATAR();
        if (userinfo.getLOGINADDRESS() != null)
            this.LOGINADDRESS = userinfo.getLOGINADDRESS();
        if (userinfo.getLOGINIP() != null)
            this.LOGINIP =  userinfo.getLOGINIP();
        if (userinfo.getISBAN() != null)
            this.ISBAN = userinfo.getISBAN();
        this.USERNUM = userinfo.getUSERNUM();
    }

    public UserBean(HashMap<String, Object> userinfo) {
        this.GUID = (String) userinfo.get("GUID");
        this.CODE = (String) userinfo.get("CODE");
        this.NAME = (String) userinfo.get("NAME");
        this.PASSWORD = (String) userinfo.get("PASSWORD");
        this.PASSWORDSALT = (String) userinfo.get("PASSWORDSALT");
        this.ROLE = (String) userinfo.get("ROLE");
        this.REMARK = (String) userinfo.get("REMARK");
        this.PHONE=(String) userinfo.get("PHONE");
        this.EMAIL=(String) userinfo.get("EMAIL");
        this.AVATAR=(String) userinfo.get("AVATAR");
        if (userinfo.get("LOGINADDRESS") != null)
            this.LOGINADDRESS = (String) userinfo.get("LOGINADDRESS");
        if (userinfo.get("LOGINIP") != null)
            this.LOGINIP = (String) userinfo.get("LOGINIP");
        if (userinfo.get("ISBAN") != null)
            this.ISBAN = (String) userinfo.get("ISBAN");
        if (userinfo.get("USERNUM") != null)
            this.USERNUM = (int) userinfo.get("USERNUM");
    }

    public UserBean(String GUID, String CODE, String NAME, String PASSWORD, String PASSWORDSALT, String ROLE, String REMARK, String LOGINADDRESS, String LOGINIP, String PHONE, String EMAIL, String AVATAR, String ISBAN,int USERNUM) {
        this.GUID = GUID;
        this.CODE = CODE;
        this.NAME = NAME;
        this.PASSWORD = PASSWORD;
        this.PASSWORDSALT = PASSWORDSALT;
        this.ROLE = ROLE;
        this.REMARK = REMARK;
        this.LOGINADDRESS = LOGINADDRESS;
        this.LOGINIP = LOGINIP;
        this.PHONE = PHONE;
        this.EMAIL = EMAIL;
        this.AVATAR = AVATAR;
        this.ISBAN = ISBAN;
        this.USERNUM = USERNUM;
    }




}