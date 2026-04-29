import SwaggerUI from 'swagger-ui-react'
import 'swagger-ui-react/swagger-ui.css'

export function APIDocument() {
    const params = new URLSearchParams(window.location.search)
    const specId = params.get('specId')?.trim() || ''
    const specUrl = params.get('specUrl')?.trim() || 'https://petstore.swagger.io/v2/swagger.json'

    let swaggerProps: any = {
        tryItOutEnabled: true,
        docExpansion: 'list'
    }

    if (specId) {
        swaggerProps.url = `/api2/openapi/${specId}`
    } else if (specUrl) {
        swaggerProps.url = specUrl
    }

    return (
        <div style={{ padding: '24px', position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, overflowY: 'auto' }}>
            <SwaggerUI {...swaggerProps} />
        </div>
    )
}