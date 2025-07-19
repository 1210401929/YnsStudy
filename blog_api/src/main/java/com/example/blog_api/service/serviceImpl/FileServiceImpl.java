package com.example.blog_api.service.serviceImpl;

import com.example.blog_api.service.FileService;
import com.example.common_api.bean.ResultBody;
import com.example.common_api.service.CallService;
import com.example.common_api.util.FunToUrlUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Component
public class FileServiceImpl implements FileService {
    @Autowired
    CallService callService;

    @Override
    public ResultBody consistencyFileCheck(String type) {
        if (type == null) {
            type = "editorImage";
        }
        //关于某类型的所有文件路径
        List<String> listUrl = getListUrlByType(type);
        List<String> fileNames = getFileNamesByUrl(listUrl);
        System.out.println("所有文件:"+fileNames);
        String url = "/opt/upload/" + type;
        File dir = new File(url);
        if (!dir.exists() || !dir.isDirectory()) {
            return ResultBody.createErrorResult("目录不存在: " + url);
        }
        File[] files = dir.listFiles();
        if (files == null) {
            return ResultBody.createErrorResult("无法读取目录: " + url);
        }
        List<String> deletedFiles = new ArrayList<>();
        for (File file : files) {
            String fileName = file.getName();
            // 如果该文件路径不在 listUrl 中，则删除
            if (!fileNames.contains(fileName)) {
                boolean deleted = file.delete();
                if (deleted) {
                    deletedFiles.add(fileName);
                }
            }
        }
        return ResultBody.createSuccessResult(deletedFiles);
    }

    /**
     * 根据url获取所有该路径下所有文件名
     *
     * @param listUrl
     * @return
     */
    private List<String> getFileNamesByUrl(List<String> listUrl) {
        List<String> fileNames = new ArrayList<>();
        for (String url : listUrl) {
            String fileName = url.substring(url.lastIndexOf("/") + 1);
            fileNames.add(fileName);
        }
        return fileNames;
    }

    /**
     * 获取某种类型  所有的文件url
     *
     * @param type editorImage|resourcesFile|userAvatar  对应  编辑器图片|资源上传文件|用户头像文件
     * @return
     */
    private List<String> getListUrlByType(String type) {

        String sql;
        ResultBody result;
        List<Map<String, String>> list;
        List<String> listUrl = new ArrayList<>();
        switch (type) {
            case "editorImage":
                sql = "select mainText as TEXT from blogInfo";
                result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
                list = (List<Map<String, String>>) result.result;
                for (Map<String, String> item : list) {
                    if (item.get("TEXT") != null) {
                        List<String> oneBlogImageUrls = analyzeHtml2ImageUrl(item.get("TEXT"));
                        listUrl.addAll(oneBlogImageUrls);
                    }
                }
                break;
            case "resourcesFile":
                sql = "select guid,fileViewUrl as URL from fileInfo";
                result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
                list = (List<Map<String, String>>) result.result;
                for (Map<String, String> item : list) {
                    if (item.get("URL") != null) {
                        listUrl.add(item.get("URL"));
                    }
                }
                break;
            case "userAvatar":
                sql = "select guid,avatar as URL from userInfo";
                result = callService.callFunOneParams(FunToUrlUtil.selectListUrl, "sql", sql);
                list = (List<Map<String, String>>) result.result;
                for (Map<String, String> item : list) {
                    if (item.get("URL") != null) {
                        listUrl.add(item.get("URL"));
                    }
                }
                break;
        }
        return listUrl;
    }

    //解析html 列出包含的图片地址
    private List<String> analyzeHtml2ImageUrl(String html) {
        List<String> srcList = new ArrayList<>();
        // 提取所有 img 标签
        Pattern imgPattern = Pattern.compile("<img\\b[^>]*?\\bsrc\\s*=\\s*\"([^\"]+)\"", Pattern.CASE_INSENSITIVE);
        Matcher matcher = imgPattern.matcher(html);
        while (matcher.find()) {
            srcList.add(matcher.group(1)); // group(1) 是括号中的内容，即 src 的值
        }
        return srcList;
    }
}
