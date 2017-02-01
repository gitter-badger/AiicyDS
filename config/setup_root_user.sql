update user_info set status=1 where uid=1; --- 设置用户状态为审核，即不用检查邮件，也可使用
update user_info set is_root=1 where uid=1; --- 把uid为1的用户，设置为root用户
update user_info set open=0 where uid=1; --- 不公开管理员的邮件

