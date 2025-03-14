import express from "express";
import cors from "cors";
import puppeteer from "puppeteer";

const app = express();
const port = process.env.PORT || 3000;

app.use(cors());
app.use(express.json());

const authMiddleware = (
  req: express.Request,
  res: express.Response,
  next: express.NextFunction
): any => {
  const token = req.query.t as string;
  console.log(process.env.TOKEN);
  if (!token || token !== process.env.TOKEN) {
    return res.status(401).json({ error: "Unauthorized" });
  }
  next();
};

app.get("/bigsnow", authMiddleware, async (req, res): Promise<void> => {
  const url =
    "https://bigsnowad.snowcloud.shop/shop/page/1E7B1BEE-0982-4F86-0F80-FC2A96F03E19";

  const selector = "h3.text-primary.mb-n1";

  try {
    const browser = await puppeteer.launch({
      headless: true,
      timeout: 60000,
      args: [
        "--no-sandbox",
        "--disable-setuid-sandbox",
        "--window-size=1920,1080",
        "--disable-dev-shm-usage",
        "--disable-gpu",
        "--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
      ],
    });
    const page = await browser.newPage();

    await page.setViewport({ width: 1920, height: 1080 });

    await page.setExtraHTTPHeaders({
      "Accept-Language": "en-US,en;q=0.9",
      Accept:
        "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
    });

    await page.goto(url, { waitUntil: "networkidle0" });

    const contents = await page.evaluate((selector) => {
      const elements = document.querySelectorAll(selector);
      return Array.from(elements).map((element) => ({
        text: element.textContent?.trim() || "",
        html: element.innerHTML,
      }));
    }, selector);

    await browser.close();

    if (contents.length === 0) {
      res.json({ error: "No matching elements found" });
    } else {
      res.json({ contents });
    }
  } catch (error) {
    console.error("Error fetching website:", error);
    res.status(500).json({ error: "Failed to fetch website" });
  }
});

const patagoniaRequestHandler =
  (url: string) =>
  async (req: express.Request, res: express.Response): Promise<void> => {
    try {
      const browser = await puppeteer.launch({
        headless: true,
        timeout: 60000,
        args: [
          "--no-sandbox",
          "--disable-setuid-sandbox",
          "--window-size=1920,1080",
          "--disable-dev-shm-usage",
          "--disable-gpu",
          "--user-agent=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
        ],
      });
      const page = await browser.newPage();

      await page.setViewport({ width: 1920, height: 1080 });

      await page.setExtraHTTPHeaders({
        "Accept-Language": "en-US,en;q=0.9",
        Accept:
          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
      });

      await page.goto(url, { waitUntil: "networkidle0" });

      const contents = await page.evaluate((selector) => {
        const elements = document.querySelectorAll(selector);
        return Array.from(elements).map((element) => ({
          text: element.textContent?.trim() || "",
          html: element.innerHTML,
        }));
      }, "span.value");

      await browser.close();

      if (contents.length === 0) {
        res.json({ error: "No matching elements found" });
      } else {
        res.json({ contents: contents[0] });
      }
    } catch (error) {
      console.error("Error fetching website:", error);
      res.status(500).json({ error: "Failed to fetch website" });
    }
  };

app.get("/patagonia-glacier", authMiddleware, (req, res) =>
  patagoniaRequestHandler(
    "https://www.patagonia.com/product/mens-jackson-glacier-down-jacket/27921.html?cgid=mens-jackets-vests-insulated"
  )(req, res)
);

app.get("/patagonia-nano-puff", authMiddleware, (req, res) =>
  patagoniaRequestHandler(
    "https://www.patagonia.com/product/womens-nano-puff-insulated-jacket/84217.html?dwvar_84217_color=BLK"
  )(req, res)
);

app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
