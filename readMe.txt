--------------------关于 启动项目--------------------------------------------------
项目启动顺序:
后端(依次执行4个spring boot项目):
1.EurekaServerApplication
2.GatewayServiceApplication
3.PubApiApplication
4.BlogApiApplication
前端:
进入yns-vue-web目录 -> 启动命令:npm run dev


--------------------关于 项目--------------------------------------------------
eureka-server:注册中心
gateway-service:网关
common_api:公共模块,类似于UserBean ResultBody类的定义以及通用实现方法,其他微服务pom文件调用该项目
pub_api:接口集成服务,用于大多数通用api
blog_api:博客相关服务  新增文章,评论等等
yns-vue-web:vue前端项目  是yns_portal的迭代  实现真正前后端分离
yns_portal:前端(已废弃)

--------------------关于 配置--------------------------------------------------
关于跨域请求 统一网关(gateway)处理
如增加子项目, 需要配置网关配置文件 和 vue项目的vite.config.js文件

资源文件映射,配置在pub_api/src/main/java/com/example/pub_api/config/StaticResourceConfig.java

项目必须依赖nginx 进行静态文件跳转,以及端口跳转
--------------------关于数据库--------------------------------------------------
userInfo =>用户表
userFollow =>用户关注列表
loginHistory =>用户登录历史表

noticeInfo =>消息通知表

fileInfo =>附件表
blogInfo =>博客文章表;
blogComment =>博客评论表;
blogGiveLike =>博客点赞收藏表

communityInfo =>社区内容表
communityComment =>社区内容评论表


建表,GUID是必须存在的字段 设置随机32位码;

---------------------其他-------------------------------------------------
因云服务器提示6379对外攻击:pub_api和blog_api的yml配置文件中的redis密码不再公开

前台静态文件访问,需要通过nginx进行转发到网关  包括所有图片和上传的资源文件

富文本编辑器的图片,路径传到
项目根目录/upload/editorImage

资源文件,路径
项目根目录/upload/resourcesFile