<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>上传文件</title>
  </head>
  <body>
    <form enctype="multipart/form-data" action="/upload" method="post">
      <input type="file" name="uploadfile" />
      <input type="hidden" name="token" value="{{.}}" />
      <input type="submit" value="upload" />
    </form>
  </body>
</html>
