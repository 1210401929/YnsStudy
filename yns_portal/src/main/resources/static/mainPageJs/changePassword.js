function changePassword() {
    event.preventDefault();
    let oldPassWord = document.getElementById("currentPassword").value;
    let newPassWord = document.getElementById("newPassword").value;
    let confirmPassword = document.getElementById("confirmPassword").value;

    let result = Y.util.ajax("/changePassWord", {oldPassWord, newPassWord, confirmPassword});
    if (result && !result.isError) {
        alert("执行成功,请重新登录!");
        parent.location.href = `/login`;
        return true;
    } else {
        alert(result.errMsg);
        return false;
    }
}