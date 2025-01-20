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
) => {
  const token = req.query.t as string;
  console.log(process.env.TOKEN);
  if (!token || token !== process.env.TOKEN) {
    return res.status(401).json({ error: "Unauthorized" });
  }
  next();
};

app.get("/fetch-website", authMiddleware, async (req, res) => {
  const url = req.query.url as string;
  if (!url) {
    return res.status(400).json({ error: "URL parameter is required" });
  }

  try {
    const browser = await puppeteer.launch({
      headless: true,
      args: ["--no-sandbox", "--disable-setuid-sandbox"],
    });
    const page = await browser.newPage();

    await page.goto(url, { waitUntil: "networkidle0" });

    const contents = await page.evaluate(() => {
      const elements = document.querySelectorAll("h3.text-primary.mb-n1");
      return Array.from(elements).map((element) => ({
        text: element.textContent?.trim() || "",
        html: element.innerHTML,
      }));
    });

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

app.listen(port, () => {
  console.log(`Server is running on port ${port}`);
});
