import SwaggerUI from 'swagger-ui-react'
import 'swagger-ui-react/swagger-ui.css'
import { useSearch } from '@tanstack/react-router'

export function APIDocument(){
    const searchParams = useSearch({
        select: (search) => ({
            specUrl: search.specUrl as string || 'https://petstore.swagger.io/v2/swagger.json',
            specId: search.specId as string || ''
        })
    })

    let swaggerProps: any = {
        tryItOutEnabled: true,
        docExpansion: 'list'
    }

    if (searchParams.specId) {
        swaggerProps.url = `/api2/openapi/${searchParams.specId}`
    } else if (searchParams.specUrl) {
        swaggerProps.url = searchParams.specUrl
    }

    return (
        <div style={{ padding: '24px', minHeight: '100vh' }}>
            <SwaggerUI {...swaggerProps} />
        </div>
    )
}