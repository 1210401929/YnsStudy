package com.example.common_api.bean;

import java.io.Serializable;

public class ResultBody implements Serializable {
    protected static final long serialVersionUID = -5241143242596509203L;
    public boolean isError = false;
    public String errMsg = "";
    public Object result = null;

    public ResultBody() {
    }

    public static ResultBody createErrorResult(String msg) {
        ResultBody rb = new ResultBody();
        rb.isError = true;
        rb.errMsg = msg;
        return rb;
    }

    public static ResultBody createSuccessResult(Object result) {
        ResultBody rb = new ResultBody();
        rb.isError = false;
        rb.result = result;
        return rb;
    }
}