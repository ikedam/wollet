# WOLlet

## 新規設置手順 (インターネット側)

※さくらインターネットのレンタルサーバーへの設置を前提にしています。

1. バイナリーのビルド

    ```
    docker compose run --rm build-wolbolt-cgi
    ```

    * public/wolbolt.cgi が作成されます。

2. `public/.htaccess` の作成

    * `.htaccess.example` をコピーして `.htaccess` を作成
    * `AuthUserFile` の部分を設置場所に合わせて指定

3. `public/wolbolt.yaml` の作成

    * `wolbolt.yaml.example` をコピーして `wolbolt.yaml` を作成
    * `secret`, `port` を設定


4. BASIC 認証用パスワードファイル (`public/.htpasswd`) の作成

    ```
    docker compose run --rm htpasswd -c public/.htpasswd username
    ```

5. public の中身をサーバーに設置

    * `*.example` は消して OK。あっても悪さはしない。
    * `wolbolt.cgi` は実行権限を設定すること。

## 新規設置手順 (ローカル側)

※mipsle の OpenWRT への設置を前提にしています。

1. バイナリーをビルド

    ```
    docker compose run --rm build-wolnut
    ```

    * `wolnut` が出来上がります。

2. `wolnut.yaml` を作成

    * `wolnut.yaml.example` をコピーして `wolnut.yaml` を作成。
    * ひととおり設定 (基本的に全項目設定が必要)

3. SCP で OpenWRT に `wolnut` と `wolnut.yaml` を設置 (以降、 /root に設置した前提で記載)

    * `wolnut` には実行権限を設定すること。

4. System > Startup > Local Startup に以下のように記載

    ```
    /root/wolnut 2>&1 | logger -t wolnut &
    ```

5. OpenWRT を再起動
