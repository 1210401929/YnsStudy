package com.example.common_api.bean;


import java.io.Serializable;
import java.util.HashMap;

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
    public String getLOGINADDRESS() {
        return LOGINADDRESS;
    }

    public void setLOGINADDRESS(String LOGINADDRESS) {
        this.LOGINADDRESS = LOGINADDRESS;
    }

    public String getLOGINIP() {
        return LOGINIP;
    }

    @Override
    public String toString() {
        return "UserBean{" +
                "GUID='" + GUID + '\'' +
                ", CODE='" + CODE + '\'' +
                ", NAME='" + NAME + '\'' +
                ", PASSWORD='" + PASSWORD + '\'' +
                ", ROLE='" + ROLE + '\'' +
                ", REMARK='" + REMARK + '\'' +
                ", LOGINADDRESS='" + LOGINADDRESS + '\'' +
                ", LOGINIP='" + LOGINIP + '\'' +
                '}';
    }

    public void setLOGINIP(String LOGINIP) {
        this.LOGINIP = LOGINIP;
    }



    public UserBean(HashMap<String, Object> userinfo) {
        this.GUID = (String) userinfo.get("GUID");
        this.CODE = (String) userinfo.get("CODE");
        this.NAME = (String) userinfo.get("NAME");
        this.PASSWORD = (String) userinfo.get("PASSWORD");
        this.PASSWORDSALT = (String) userinfo.get("PASSWORDSALT");
        this.ROLE = (String) userinfo.get("ROLE");
        this.REMARK = (String) userinfo.get("REMARK");
        if (userinfo.get("LOGINADDRESS") != null)
            this.LOGINADDRESS = (String) userinfo.get("LOGINADDRESS");
        if (userinfo.get("LOGINIP") != null)
            this.LOGINADDRESS = (String) userinfo.get("LOGINIP");
    }

    public UserBean(String GUID, String CODE, String NAME, String PASSWORD,  String PASSWORDSALT ,String ROLE, String REMARK) {
        this.GUID = GUID;
        this.CODE = CODE;
        this.NAME = NAME;
        this.PASSWORD = PASSWORD;
        this.PASSWORDSALT = PASSWORDSALT;
        this.ROLE = ROLE;
        this.REMARK = REMARK;
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

}