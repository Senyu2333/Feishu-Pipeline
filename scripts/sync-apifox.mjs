#!/usr/bin/env node
import fs from 'node:fs/promises'

const specPath = process.argv[2] || process.env.SWAGGER_JSON_PATH || 'apps/api-go/docs/swagger.json'
const projectId = process.env.APIFOX_PROJECT_ID
const accessToken = process.env.APIFOX_ACCESS_TOKEN
const endpoint = process.env.APIFOX_IMPORT_ENDPOINT || 'https://api.apifox.com/v1/projects/openapi/import'

if (!projectId) {
  console.error('Missing APIFOX_PROJECT_ID')
  process.exit(1)
}

if (!accessToken) {
  console.error('Missing APIFOX_ACCESS_TOKEN')
  process.exit(1)
}

const raw = await fs.readFile(specPath, 'utf8')
const openapiSpec = JSON.parse(raw)

const payload = {
  project_id: projectId,
  openapi_spec: openapiSpec,
  options: {
    endpoint_overwrite_behavior: process.env.APIFOX_ENDPOINT_OVERWRITE_BEHAVIOR || 'coverUnmatchedResources',
    schema_overwrite_behavior: process.env.APIFOX_SCHEMA_OVERWRITE_BEHAVIOR || 'COVER_EXISTING',
    update_folder_of_changed_endpoint: process.env.APIFOX_UPDATE_FOLDER_OF_CHANGED_ENDPOINT !== 'false',
    prepend_base_path: process.env.APIFOX_PREPEND_BASE_PATH === 'true',
  },
}

if (process.env.APIFOX_TARGET_ENDPOINT_FOLDER_ID) {
  payload.options.target_endpoint_folder_id = Number(process.env.APIFOX_TARGET_ENDPOINT_FOLDER_ID)
}
if (process.env.APIFOX_TARGET_SCHEMA_FOLDER_ID) {
  payload.options.target_schema_folder_id = Number(process.env.APIFOX_TARGET_SCHEMA_FOLDER_ID)
}

const response = await fetch(endpoint, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${accessToken}`,
  },
  body: JSON.stringify(payload),
})

const text = await response.text()
let data
try {
  data = JSON.parse(text)
} catch {
  data = text
}

if (!response.ok) {
  console.error('Apifox import failed')
  console.error(JSON.stringify(data, null, 2))
  process.exit(1)
}

console.log(JSON.stringify(data, null, 2))
