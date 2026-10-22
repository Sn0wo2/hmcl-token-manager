# hmcl-token-manager

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Sn0wo2/hmcl-token-manager)

> 帮助你导入 / 导出 [HMCL](https://github.com/HMCL-dev/HMCL) Minecraft 账户`refresh_token`和`access_token`的小工具

> [!TIP]
> 如果你需要 Account Share 给朋友但是不想登录 Microsoft 账户的话应该会对你有帮助

### 引言

HMCL 在 [v3.16.0.347](https://docs.hmcl.net/changelog/dev.html#HMCL-3.16.0.347) 引入了 hmcl-obfuscated-v1 用于混淆`access_token`和`refresh_token`的存储, 并一起更新了一堆数据模型
在此之前我们只需要直接去 `HMCL_USER_HOME/.hmcl/` 找到 `accounts.json` 就可以导出账户信息. 现在账户数据经过混淆且结构有所变化, 手动导入导出非常坐牢