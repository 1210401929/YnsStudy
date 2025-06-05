package com.example.blog_api.Bean;


public class BlogBean {
    private String GUID;
    private String BLOG_TITLE;
    private String BLOG_TYPE;
    private String MAINTEXT;
    private String USERCODE;
    private String USERNAME;

    public BlogBean(String GUID, String BLOG_TITLE, String BLOG_TYPE, String MAINTEXT, String USERCODE, String USERNAME) {
        this.GUID = GUID;
        this.BLOG_TITLE = BLOG_TITLE;
        this.BLOG_TYPE = BLOG_TYPE;
        this.MAINTEXT = MAINTEXT;
        this.USERCODE = USERCODE;
        this.USERNAME = USERNAME;
    }

    public String getGUID() {
        return GUID;
    }

    public void setGUID(String GUID) {
        this.GUID = GUID;
    }

    public String getBLOG_TITLE() {
        return BLOG_TITLE;
    }

    public void setBLOG_TITLE(String BLOG_TITLE) {
        this.BLOG_TITLE = BLOG_TITLE;
    }

    public String getBLOG_TYPE() {
        return BLOG_TYPE;
    }

    public void setBLOG_TYPE(String BLOG_TYPE) {
        this.BLOG_TYPE = BLOG_TYPE;
    }

    public String getMAINTEXT() {
        return MAINTEXT;
    }

    public void setMAINTEXT(String MAINTEXT) {
        this.MAINTEXT = MAINTEXT;
    }

    public String getUSERCODE() {
        return USERCODE;
    }

    public void setUSERCODE(String USERCODE) {
        this.USERCODE = USERCODE;
    }

    public String getUSERNAME() {
        return USERNAME;
    }

    public void setUSERNAME(String USERNAME) {
        this.USERNAME = USERNAME;
    }
}
