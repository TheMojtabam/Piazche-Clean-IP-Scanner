# Shodan Integration

پیازچه می‌تونه از Shodan API برای پیدا کردن IP های non-Cloudflare که **پشت Cloudflare** هستن استفاده کنه.

## چرا؟

این IP ها رنج‌های اصلی CF نیستن (مثلاً `104.x.x.x`) ولی SSL certificate یا هدرهای Cloudflare دارن — یعنی شرکت‌های دیگه‌ای که از CF CDN استفاده می‌کنن. این IP ها معمولاً:
- کمتر فیلتر می‌شن
- محدودیت کمتری دارن
- می‌تونن با config شما جواب بدن

## تنظیم config.json

```json
"shodan": {
  "apiKey": "YOUR_SHODAN_API_KEY",
  "mode": "scan",
  "useDefaultQuery": true,
  "query": "",
  "pages": 2,
  "excludeCFRanges": true,
  "saveHarvestedIPs": "results/shodan_ips.txt",
  "appendToExisting": false
}
```

## حالت‌های `mode`

| حالت | توضیح |
|------|-------|
| `off` | غیرفعال (پیش‌فرض) — از فایل ipv4.txt اسکن کن |
| `harvest` | فقط IP جمع کن از Shodan، ذخیره کن، اسکن نکن |
| `scan` | IP جمع کن از Shodan و بلافاصله اسکن کن |
| `both` | IP جمع کن، ذخیره کن، **و** اسکن کن |

## اجرا از command line

```bash
# فقط جمع‌آوری IP
./piyazche -c config.json --shodan-mode harvest --shodan-key YOUR_KEY

# جمع‌آوری + اسکن
./piyazche -c config.json --shodan-mode scan --shodan-pages 3

# جمع‌آوری 5 صفحه (=500 IP)
./piyazche -c config.json --shodan-mode both --shodan-pages 5
```

## هزینه Query Credit

هر صفحه = 100 IP = **1 query credit** مصرف می‌کنه.

- Plan رایگان Shodan: محدود
- Plan پولی: بیشتر

با `--shodan-pages 1` شروع کن.

## کوئری پیش‌فرض

```
ssl:"Cloudflare Inc ECC CA" port:443 -net:173.245.48.0/20 -net:103.21.244.0/22 ...
```

این کوئری IP هایی پیدا می‌کنه که:
- SSL cert از Cloudflare دارن
- رنج اصلی CF **نیستن**
- پورت 443 باز دارن

## کوئری سفارشی

```json
"query": "http.title:\"Cloudflare\" port:443 country:US",
"useDefaultQuery": false
```
