package com.example.blog_api.service;

import com.example.common_api.bean.ResultBody;
import org.springframework.stereotype.Service;

import java.util.HashMap;

@Service
public interface FriendLinkService {

    ResultBody addFriendLink(HashMap<String,Object> friendLink);
    ResultBody updateFriendLink(HashMap<String,Object> friendLink);
    ResultBody deleteFriendLink(String friendLinkId);
    ResultBody getFriendLink(String friendLinkId);
    ResultBody getFriendLinks();
}
