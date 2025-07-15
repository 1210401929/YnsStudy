package com.example.pub_api.service.serviceImpl;

import com.example.common_api.bean.ResultBody;
import com.example.common_api.config.CommonFileCfg;
import com.example.pub_api.service.UploadService;
import org.springframework.stereotype.Component;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

@Component
public class UploadServiceImpl implements UploadService {
    CommonFileCfg commonCfg = new CommonFileCfg();

    @Override
    public ResultBody uploadFile(MultipartFile file, String spliceUrl) throws IOException {
        if (file.isEmpty()) {
            throw new IllegalArgumentException("上传文件不能为空");
        }
        String baseUploadPath = getUploadPath();  // 物理路径根目录

        String relativeDir = ""; //前端传过来的附加路径
        if (spliceUrl != null) {
            relativeDir = spliceUrl;  // 相对目录，比如 "editorImage/"
        }

        String fullUploadPath = baseUploadPath + relativeDir;

        // 创建上传目录（如果不存在）
        File dir = new File(fullUploadPath);
        if (!dir.exists()) {
            dir.mkdirs();
        }

        // 原始文件名
        String originalFilename = file.getOriginalFilename();
        // 生成唯一文件名
        String ext = originalFilename.substring(originalFilename.lastIndexOf('.'));
        String fileName = UUID.randomUUID() + ext;
        File dest = new File(dir, fileName);
        file.transferTo(dest);
        //获取网关IP端口号
        String gateWayIpPort = getIsDev() ? commonCfg.getGateWayDevUrl() : commonCfg.getGateWayProduceUrl();
        //文件用来访问的路径  这里返回的是映射路径(映射文件在../../config/StaticResourceConfig.java)，用于前端访问
        //String fileViewUrl = gateWayIpPort + commonCfg.getMappingStaticUrl() + relativeDir + "/" + fileName;
        String fileViewUrl = commonCfg.getMappingStaticUrl() + relativeDir + "/" + fileName;
        //文件真实存储的路径
        String fileRealUrl = fullUploadPath + "/" + fileName;
        // 返回数据
        Map<String, String> fileInfo = new HashMap<>();
        fileInfo.put("originalFileName", originalFilename);
        fileInfo.put("savedFileName", fileName);
        fileInfo.put("fileViewUrl", fileViewUrl);
        fileInfo.put("fileRealUrl", fileRealUrl);
        System.out.println("已上传文件,文件信息:" + fileInfo);
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
