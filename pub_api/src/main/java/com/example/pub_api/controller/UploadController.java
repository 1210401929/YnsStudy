package com.example.pub_api.controller;

import com.example.common_api.bean.ResultBody;
import com.example.pub_api.service.UploadService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.util.List;
import java.util.Map;

@Controller
@RequestMapping("pub-api/upload")
public class UploadController {

    @Autowired
    UploadService uploadService;

    @PostMapping("/uploadFile")
    @ResponseBody
    public ResultBody uploadFile(@RequestParam("file") MultipartFile file,
                                 @RequestParam(value = "spliceUrl", required = false) String spliceUrl) throws IOException {
        return uploadService.uploadFile(file, spliceUrl);
    }

    @PostMapping("/deleteFileByUrl")
    @ResponseBody
    public ResultBody deleteFileByUrl(@RequestBody Map<String,String> params){
        String url = params.get("url");
        return uploadService.deleteFileByUrl(url);
    }
    @PostMapping("/deleteFileByUrls")
    @ResponseBody
    public ResultBody deleteFileByUrls(@RequestBody Map<String,Object> params){
        List<String> urls =(List<String>) params.get("urls");
        return uploadService.deleteFileByUrls(urls);
    }
}
