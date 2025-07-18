package com.example.blog_api.controller;

import com.example.blog_api.Bean.BlogBean;
import com.example.blog_api.service.BlogService;
import com.example.common_api.bean.ResultBody;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseBody;

import javax.servlet.http.HttpSession;
import java.util.Map;

@Controller
@RequestMapping("blog-api/blog")
public class BlogController {

    @Autowired
    BlogService blogService;

    @RequestMapping("/addBlog")
    @ResponseBody
    public ResultBody addBlog(@RequestBody Map<String, Object> params) {
        Map<String, String> blogMap = (Map<String, String>) params.get("blogContent");
        BlogBean blogBean = new BlogBean(blogMap.get("GUID"), blogMap.get("BLOG_TITLE"), blogMap.get("BLOG_TYPE"), blogMap.get("MAINTEXT"), blogMap.get("USERCODE"), blogMap.get("USERNAME"));
        return blogService.addBlog(blogBean);
    }

    @RequestMapping("/deleteBlog")
    @ResponseBody
    public ResultBody deleteBlog(@RequestBody Map<String, String> params) {
        String guid = params.get("guid");
        return blogService.deleteBlog(guid);
    }

    @RequestMapping("/updateBlog")
    @ResponseBody
    public ResultBody updateBlog(@RequestBody Map<String, String> params) {
        String guid = params.get("guid");
        String blog_type = params.get("blog_type");
        String title = params.get("title");
        String content = params.get("content");
        return blogService.updateBlog(guid, title, content,blog_type);
    }

    @RequestMapping("/getUserBlog")
    @ResponseBody
    public ResultBody getUserBlog(HttpSession session) {
        return blogService.getUserBlog(session);
    }

    @RequestMapping("/getBlog")
    @ResponseBody
    public ResultBody getUserBlog(@RequestBody Map<String, String> params, HttpSession session) {
        String blogId = params.get("blogId");
        return blogService.getBlog(session, blogId);
    }

    @RequestMapping("/getAllBlog")
    @ResponseBody
    public ResultBody getAllBlog(@RequestBody Map<String, Object> params) {
        int page = (int)params.get("page");
        int pageSize = (int)params.get("pageSize");
        String keyword = (String) params.get("keyword");
        return blogService.getAllBlog(page, pageSize, keyword);
    }

    @RequestMapping("/addComment")
    @ResponseBody
    public ResultBody addComment(@RequestBody Map<String, Object> params, HttpSession session) {
        Map<String, String> blogComment = (Map<String, String>) params.get("blogComment");
        return blogService.addComment(blogComment);
    }

    @RequestMapping("/getComment")
    @ResponseBody
    public ResultBody getComment(@RequestBody Map<String, String> params) {
        String blogId = params.get("blogId");
        return blogService.getComment(blogId);
    }

    @RequestMapping("/deleteComment")
    @ResponseBody
    public ResultBody deleteComment(@RequestBody Map<String, Object> params) {
        String blogGuid = (String) params.get("blogGuid");
        return blogService.deleteComment(blogGuid);
    }

    @RequestMapping("/giveLikeBlog")
    @ResponseBody
    public ResultBody giveLikeBlog(@RequestBody Map<String, Object> params,HttpSession session) {
        String blogId = (String) params.get("blogId");
        return blogService.giveLikeBlog(blogId,session);
    }
    @RequestMapping("/noGiveLikeBlog")
    @ResponseBody
    public ResultBody noGiveLikeBlog(@RequestBody Map<String, Object> params,HttpSession session) {
        String blogId = (String) params.get("blogId");
        return blogService.noGiveLikeBlog(blogId,session);
    }
    @RequestMapping("/getGiveLikeByBlogId")
    @ResponseBody
    public ResultBody getGiveLikeByBlogId(@RequestBody Map<String, Object> params) {
        String blogId = (String) params.get("blogId");
        return blogService.getGiveLikeByBlogId(blogId);
    }

    @RequestMapping("/getLikeAndCollectByBlogId")
    @ResponseBody
    public ResultBody getLikeAndCollectByBlogId(@RequestBody Map<String, Object> params) {
        String blogId = (String) params.get("blogId");
        return blogService.getLikeAndCollectByBlogId(blogId);
    }
    @RequestMapping("/getLikeAndCollectByUserCode")
    @ResponseBody
    public ResultBody getLikeAndCollectByUserCode(@RequestBody Map<String, Object> params) {
        String userCode = (String) params.get("userCode");
        return blogService.getLikeAndCollectByUserCode(userCode);
    }
    @RequestMapping("/collectBlog")
    @ResponseBody
    public ResultBody collectBlog(@RequestBody Map<String, Object> params,HttpSession session) {
        String blogId = (String) params.get("blogId");
        return blogService.collectBlog(blogId,session);
    }

    @RequestMapping("/noCollectBlog")
    @ResponseBody
    public ResultBody noCollectBlog(@RequestBody Map<String, Object> params,HttpSession session) {
        String blogId = (String) params.get("blogId");
        return blogService.noCollectBlog(blogId,session);
    }
}
