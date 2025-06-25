package com.example.common_api.config;

public class CommonFileCfg {
    //开发环境 gateway网关ip端口
    private String gateWayDevUrl = "http://localhost:8889";
    //生产环境 gateway网关ip端口
    private String gateWayProduceUrl = "http://152.136.247.100:8889";
    //windows环境 文件上传的地址
    private String uploadFileWinUrl = System.getProperty("user.dir") + "/upload/";
    //linux环境 文件上传的地址
    private String uploadFileLinuxUrl = "/opt/upload/";
    //映射前端访问的地址
    private String mappingStaticUrl = "/uploadFile/";

    public String getGateWayDevUrl() {
        return gateWayDevUrl;
    }

    public void setGateWayDevUrl(String gateWayDevUrl) {
        this.gateWayDevUrl = gateWayDevUrl;
    }

    public String getGateWayProduceUrl() {
        return gateWayProduceUrl;
    }

    public void setGateWayProduceUrl(String gateWayProduceUrl) {
        this.gateWayProduceUrl = gateWayProduceUrl;
    }

    public String getUploadFileWinUrl() {
        return uploadFileWinUrl;
    }

    public void setUploadFileWinUrl(String uploadFileWinUrl) {
        this.uploadFileWinUrl = uploadFileWinUrl;
    }

    public String getUploadFileLinuxUrl() {
        return uploadFileLinuxUrl;
    }

    public void setUploadFileLinuxUrl(String uploadFileLinuxUrl) {
        this.uploadFileLinuxUrl = uploadFileLinuxUrl;
    }

    public String getMappingStaticUrl() {
        return mappingStaticUrl;
    }

    public void setMappingStaticUrl(String mappingStaticUrl) {
        this.mappingStaticUrl = mappingStaticUrl;
    }
}
