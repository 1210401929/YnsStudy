package com.example.pub_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;
import org.springframework.web.multipart.MultipartFile;

import java.io.IOException;
import java.util.List;

@Service
public interface UploadService {
    ResultBody uploadFile(MultipartFile file, String spliceUrl) throws IOException;

    ResultBody deleteFileByUrl(String urlPath);
    ResultBody deleteFileByUrls(List<String> urls);
}
