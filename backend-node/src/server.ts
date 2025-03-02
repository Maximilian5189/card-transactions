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
  if (!url) {
    res.status(400).json({ error: "URL parameter is required" });
  }

  try {
    const browser = await puppeteer.launch({
      headless: false,
      args: ["--no-sandbox", "--disable-setuid-sandbox"],
    });
    const page = await browser.newPage();

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

const patagoniaRequestHandler = async (req, res): Promise<void> => {
  const url =
    "https://www.patagonia.com/product/mens-jackson-glacier-down-jacket/27921.html?cgid=mens-jackets-vests-insulated";

  const selector = "span.value";
  if (!url) {
    res.status(400).json({ error: "URL parameter is required" });
  }

  try {
    const browser = await puppeteer.launch({
      headless: false,
      args: ["--no-sandbox", "--disable-setuid-sandbox"],
    });
    const page = await browser.newPage();

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
};

app.get("/patagonia-glacier", authMiddleware, patagoniaRequestHandler);

app.get("/patagonia-nano-puff", authMiddleware, patagoniaRequestHandler);

app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
