package com.example.pub_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.config.CommonFileCfg;
import com.example.common_api.config.PublicCfg;
import com.example.pub_api.service.UploadService;
import org.springframework.stereotype.Component;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.util.*;

@Component
public class UploadServiceImpl implements UploadService {
    CommonFileCfg commonCfg = new CommonFileCfg();

    @Override
    public ResultBody uploadFile(MultipartFile file, String spliceUrl) throws IOException {
        if (file.isEmpty()) {
            throw new IllegalArgumentException("上传文件不能为空");
        }

        //后缀校验逻辑
        String originalFilename = file.getOriginalFilename();
        if (originalFilename == null || !originalFilename.contains(".")) {
            throw new IllegalArgumentException("非法文件名，缺少扩展名");
        }
        // 提取后缀并转为小写
        String ext = originalFilename.substring(originalFilename.lastIndexOf('.')).toLowerCase();
        // 定义允许的白名单（建议根据业务需求调整）
        Set<String> allowedExtensions = PublicCfg.allowedExtensions;
        //判断allowedExtensions数组是否包含ext后缀  如果不包含->报错
        if (!allowedExtensions.contains(ext)) {
            return ResultBody.createErrorResult("不支持的文件类型：" + ext);
        }
        // -----------------------
        String relativeDir = (spliceUrl != null) ? spliceUrl : "";
        //判断是否包含..这种非法路径
        if (relativeDir.contains("..")) {
            return ResultBody.createErrorResult("目录路径包含非法字符");
        }
        String baseUploadPath = getUploadPath();
        String fullUploadPath = baseUploadPath + relativeDir;

        File dir = new File(fullUploadPath);
        if (!dir.exists()) {
            dir.mkdirs();
        }

        // 生成唯一文件名
        String fileName = UUID.randomUUID().toString() + ext;
        File dest = new File(dir, fileName);
        file.transferTo(dest);

        // 映射路径处理
        String fileViewUrl = commonCfg.getMappingStaticUrl() + relativeDir + "/" + fileName;
        String fileRealUrl = fullUploadPath + "/" + fileName;

        Map<String, String> fileInfo = new HashMap<>();
        fileInfo.put("originalFileName", originalFilename);
        fileInfo.put("savedFileName", fileName);
        fileInfo.put("fileViewUrl", fileViewUrl);
        fileInfo.put("fileRealUrl", fileRealUrl);

        return ResultBody.createSuccessResult(fileInfo);
    }

    /**
     * 删除文件，传入映射路径，如 "/uploadFile/editorImage/xxxx.png"
     */
    public ResultBody deleteFileByUrl(String urlPath) {
        String staticUrl = commonCfg.getMappingStaticUrl();
        if (urlPath == null || urlPath.indexOf(staticUrl) == -1) {
            return ResultBody.createErrorResult("路径不合法!");
        }
        String baseUploadPath = getUploadPath();  // 物理路径根目录
        // 去掉映射前缀
        int index = urlPath.indexOf(staticUrl);
        String relativePath = urlPath.substring(index + staticUrl.length());
        // 拼接成磁盘真实路径
        File file = new File(baseUploadPath + relativePath);
        if (file.exists()) {
            file.delete();
            return ResultBody.createSuccessResult("删除成功!");
        } else {
            return ResultBody.createErrorResult("文件或路径不存在!");
        }
    }

    /**
     * 多文件删除
     *
     * @param urls
     * @return
     */
    @Override
    public ResultBody deleteFileByUrls(List<String> urls) {
        for (String urlPath : urls) {
            ResultBody result = deleteFileByUrl(urlPath);
            if (result == null || result.isError) {
                return result;
            }
        }
        return ResultBody.createSuccessResult("删除成功!");
    }

    // 获取存储的文件路径
    private String getUploadPath() {
        return getIsDev() ? commonCfg.getUploadFileWinUrl() : commonCfg.getUploadFileLinuxUrl();
    }

    // 获取当前环境是否是开发环境  现在是根据运行环境是windows还是linux
    private Boolean getIsDev() {
        String os = System.getProperty("os.name").toLowerCase();
        if (os.contains("win")) {
            System.out.println("获取当前环境: 开发环境");
            return true;
        } else {
            System.out.println("获取当前环境: 生产环境");
            return false;
        }
    }
}
