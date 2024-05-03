import Fastify from "fastify"
import { promises as fs } from "fs"
import path from "path"
import process from "process"
import { authenticate } from "@google-cloud/local-auth"
import { google } from "googleapis"
const fastify = Fastify({
  logger: true,
})

const SCOPES = ["https://www.googleapis.com/auth/gmail.readonly"]
const TOKEN_PATH = path.join(process.cwd(), "token.json")
const CREDENTIALS_PATH = path.join(process.cwd(), "credentials.json")
const getClient = async () => {
  let client
  try {
    const content = await fs.readFile(TOKEN_PATH)
    const credentials = JSON.parse(content.toString())
    client = google.auth.fromJSON(credentials)
  } catch (err) {
    fastify.log.error(err)
    return null
  }
  if (client) {
    return client
  }
  client = await authenticate({
    scopes: SCOPES,
    keyfilePath: CREDENTIALS_PATH,
  })
  if (client.credentials) {
    const content = await fs.readFile(CREDENTIALS_PATH)
    const keys = JSON.parse(content?.toString())
    const key = keys.installed || keys.web
    const payload = JSON.stringify({
      type: "authorized_user",
      client_id: key.client_id,
      client_secret: key.client_secret,
      refresh_token: client.credentials.refresh_token,
    })
    await fs.writeFile(TOKEN_PATH, payload)
  }
  return client
}

async function getEmails(client: any) {
  const gmail = google.gmail({ version: "v1", auth: client })
  const res = await gmail.users.messages.list({
    userId: "me",
    maxResults: 10,
  })

  let messages: any = []
  for (const message of res.data.messages || []) {
    let messageContent
    if (message.id) {
      messageContent = await gmail.users.messages.get({
        id: message.id,
        userId: "me",
      })
    }

    messageContent?.data.payload?.headers?.forEach((h: any) => {
      if (h?.name === "Subject") {
        messages.push(h.value)
      }
    })
  }

  return messages
}

async function listLabels(client: any) {
  const gmail = google.gmail({ version: "v1", auth: client })
  const res = await gmail.users.labels.list({
    userId: "me",
  })
  const labels = res.data.labels
  if (!labels || labels.length === 0) {
    console.log("No labels found.")
    return
  }

  return labels
}

fastify.get("/", async (request, reply) => {
  const client = await getClient()
  const labels = await listLabels(client)
  const messages = await getEmails(client)

  return { messages }
})

const start = async () => {
  try {
    await fastify.listen({ port: 3000 })
  } catch (err) {
    fastify.log.error(err)
    process.exit(1)
  }
}
start()
