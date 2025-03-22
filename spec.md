* このシステムは WOLlet (wollet) という名前で、Web 経由で LAN 内のデバイスに Wake On LAN のマジックパケットを投げるための golang 実装のシステムです。
* WOLlet は 2 つのコンポーネントから構成されます:
    * WOLbolt: グローバル IP の記録と、グローバルIPに対してUDPパケットを送るためのCGI。
    * WOLnut: UDPパケットを受けて特定の MAC アドレスに対して Wake On LAN のマジックパケットを送信する。また、定期的に WOLbot にアクセスしてグローバル IP を記録させる。
* WOLbolt / WOLnut 共通の仕様
    * ログ出力には `pkg/log` モジュールを使用する。
        * 初期の実装では `Info()` のみが提供されていますが、 `Debug()`, `Warn()`, `Error()` も追加してください。
* WOLbolt は以下の仕様です:
    * CGI アプリケーション。
    * 以下の構成にする:
        * `cmd/wolbolt-cgi` を起動エントリーポイントとする。
            * 設定ファイルの読み込みと、メインロジックの呼び出し・実行を行う。
        * `pkg/wolbolt` をメインロジックとする。
    * 実行バイナリーと同じディレクトリーにある `wolbolt.yaml` から以下の設定を読み込む:
        * `secret`: WOLnut との通信に使用するパスワード
        * `port`: WOLnut 向けの通信に使用する UDP ポート番号
        * `count`: WOLnut へのパケットの送信回数
        * `interval_secs`: WOLnut へのパケット送信時の間隔 (秒)。float64 で指定できる。
        * `pingfile`: IP アドレスを記録するファイルのパス。指定がない場合、実行バイナリーと同じディレクトリーの `ping.txt` とする。
        * `logfile`: 動作ログを記録するファイルのパス。指定がない場合、実行バイナリーと同じディレクトリーの `ping.log` とする。
    * CGI に渡される追加パスに対して以下のように動作する:
        * POST /ping
            * リクエストボディとして `secret` を受け付ける。
            * `secret` が一致する場合、以下の処理を行い、レスポンスボディとして固定の文字列「OK」を返す。
                * `pingfile` に、1行目にアクセス元の IP アドレス、2行目にアクセスされた時刻を UTC で記録する。
                    * ファイルの更新はプロセスIDを使用した一時ファイルに出力して出力が完了したらファイルの置き換えを行う手順で実施する。
            * `secret` が一致しない場合、なんの処理も行わずにレスポンスボディとして固定の文字列「OK」を返す。
            * 以下の場合に `logfile` に動作ログとして時刻、アクセス元のIPアドレス、メッセージを出力する:
                * `secret` が一致しない場合
                * `pingfile` に記録されている IP アドレスが変化した場合。元の IP アドレスと新しい IP アドレスを記録する。
        * POST /wol
            * `pingfile` に記録されている IP アドレス、`port` で指定された UDP ポートに対して、 `secret` を本文とする UDP パケットを送り、固定の文字列「OK」を返す。
                * `interval_secs` で指定された間隔で `count` 回送信する。
        * GET /
            * 固定の文字列「OK」を返す。
        * その他
            * 404 Not Found を返す。
* WOLnut は以下の仕様です:
    * CLI アプリケーション。
    * 以下の構成にする:
        * `cmd/wolnut` を起動エントリーポイントとする。
            * 設定ファイルの読み込みと、メインロジックの呼び出し・実行を行う。
        * `pkg/wolnut` をメインロジックとする。
    * 実行バイナリーと同じディレクトリーにある `wolnut.yaml` から以下の設定を読み込む:
        * `secret`: WOLbolt との通信に使用するパスワード
        * `target`: Wake on LAN のマジックパケットを送信する対象の MAC アドレス
        * `port`: WOLbolt からの通信を受け付ける UDP ポート番号
        * `ping`
            * `url`: wolbolt の `/ping` を呼び出すための URL
            * `interval_secs`: wolbolt の `/ping` を呼び出す間隔 (秒)
            * `basic_user`: BASIC 認証のユーザー名
            * `basic_pass`: BASIC 認証のパスワード
    * 起動すると以下の3つの動作をする:
        * `ping.interval_secs` で指定した間隔で、 `url` へ POST アクセスを行う。
            * `basic_user` と `basic_pass` を使用した認証ヘッダーを設定する。
            * リクエストボディは `secret` の値を設定する。
        * `port` で指定される UDP ポートをリッスンする。
            * パケットが届いたらパケット内容が `secret` と一致するか確認する。
            * 一致する場合、 `target` で指定される MAC アドレスへ Wake On LAN のマジックパケットを送り、Info レベルでログに記録する。
            * 一致しない場合、Warn レベルでログを出力する。
        * `SIGHUP`、 `SIGTERM`、 `SIGINT` シグナルを受けたらアプリケーションを終了する。
* public 以下に Web サイトへの設置時のテンプレートを作成する。
    * `index.html` を作成する。
    * 同一ディレクトリーに、 `cmd/wolbolt-cgi` をビルドした `wolbolt.cgi` が設置され CGI として呼び出せる前提とする。(テンプレートとしてはこのファイルは含めなくて良い)
    * `index.html` には「起動」ボタンが有り、これを押下すると `wolbolt.cgi` の `POST /wol` を呼び出す。
    * `.htaccess` を作成し、以下の設定を行う:
        * BASIC 認証
            * 同一ディレクトリーに `.htpasswd` ファイルがあるものとする。
        * 以下のファイルへのアクセス拒否:
            * `*.yaml`
            * `*.log`
            * `*.txt`
            * `.ht*`
* docker-compose.yaml を作成し、以下の機能を提供する:
    * `docker compose up` で、 http://localhost:8080/wolbolt/ にアクセスすると WOLbolt の動作を試せるコンテナーを起動する。
    * `docker compose run --rm build-wolbolt-cgi` で `public/wolbolt.cgi` をビルドする。
        * ビルドオプションとして `GOOS=freebsd`、`GOARCH=amd64` を指定する。
    * `docker compose run --rm build-wolnut` で `wolnut` をビルドする。
        * ビルドオプションとして `GOARCH=mipsle`、 `GOMIPS=softfloat`、 `-ldflags="-s -w"` を指定する。
    * `docker compose run --rm htpasswd` で `public/.htpasswd` を作成/更新できるようにする。

