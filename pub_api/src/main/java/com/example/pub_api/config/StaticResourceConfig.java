package com.example.pub_api.config;

import com.example.common_api.config.CommonFileCfg;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.ResourceHandlerRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class StaticResourceConfig implements WebMvcConfigurer {
    @Override
    //静态文件映射
    public void addResourceHandlers(ResourceHandlerRegistry registry) {
        String uploadPath;
        CommonFileCfg commonCfg = new CommonFileCfg();
        String os = System.getProperty("os.name").toLowerCase();
        //如果是window,映射到项目根目录
        if (os.contains("win")) {
            uploadPath = "file:" + commonCfg.getUploadFileWinUrl();
            //如果是linux  映射到/opt/upload
        } else {
            uploadPath = "file:" + commonCfg.getUploadFileLinuxUrl();
        }
        System.out.println("上传文件的路径为:" + uploadPath);
        //前端访问的路径   /uploadFile/**
        String resourceHandlerUrl = commonCfg.getMappingStaticUrl() + "**";
        System.out.println("前端访问的路径:" + resourceHandlerUrl);
        registry.addResourceHandler(resourceHandlerUrl)//    /uploadFile/
                .addResourceLocations(uploadPath)//    /opt/upload/
                .setCachePeriod(3600)
                .resourceChain(true);
    }
}
