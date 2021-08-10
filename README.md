# syncfollows

ツイッターのフォローユーザーと特定リストを同期するやつ

home_timeline API の rate limit が終わっているので。

```docker-compose
version: '3.8'

services:
  main:
    image: ghcr.io/suzutan/syncfollows:latest
    restart: always
    environment:
      CK: CK_HERE
      CS: CS_HERE
      AT: AT_HERE
      ATS: ATS_HERE
      LIST_ID: LIST_ID_HERE
```
