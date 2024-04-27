# E621 Downloader
### CLI tool for batch downloading posts and pools

The program accepts several flags:
-----
|Flags|Default value|Description|
|----|:----:|----|
|`--poolID`|`0`|Pool ID to download. Can be found in the URL of the pool as the last part of it. The arguments is not neccessary when the `--scrapPosts` flag is set to true.|
|`--wait`|`5`|Wait time between downloads to avoid being blocked by IP address.|
|`--scrapPosts`|`false`|If set to true the program will start download the posts by the search tags that are defined with the `--pTags` flag which are required to be set. Otherwise it will start download the pool.|
|`--pTags`|`""`|The search tags, that are used for scraping specific posts when the `--scrapPosts` flag is set to true. The tags are comma separated.|
|`--maxPostPages`|`0`|The maximum amount of pages with posts or in pool to download. When set to 0 the program will download all pages.|
|`--out`|`"./defOut"`|The path to the directory where downloaded files will be put.|
|`--proxy`|`""`|The URL of proxy connection. Can be used with schema such as `socks5://127.0.0.1:1080` or `http://172.16.0.1:8080`.|
|`--dbPath`|`"./defOut/downloaded.db"`|Path to database that stores your download history for deduplication and metadata storing.|

## Examples
To download pool just enter the next command:

`go run . --poolID 23886`

It will download the pool with ID 23886 and put all downloaded files into default folder.

To download the posts with specific tags you can just enter:

`go run . --scrapPosts true --pTags "loona (helluva boss),rating:safe" --maxPagesPosts 4 --out ./Loona`

It will download first four pages with these tags and put all downloaded into `./Loona` directory.