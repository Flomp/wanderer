import { createServer, get } from "node:http";
import { setResponse } from "@sveltejs/kit/node";

// Exercise the adapter's actual HTTP framing: Response.json() alone does not
// detect a stale Content-Length when a response body has been enlarged.
export async function responseOverHttp(response: Response): Promise<Response> {
    const server = createServer((_request, outgoing) => {
        void setResponse(outgoing, response).catch(error => outgoing.destroy(error));
    });
    try {
        await new Promise<void>((resolve, reject) => {
            server.once("error", reject);
            server.listen(0, "127.0.0.1", resolve);
        });
        const address = server.address();
        if (!address || typeof address === "string") throw new Error("missing HTTP listener");
        return await new Promise<Response>((resolve, reject) => {
            const request = get({ hostname: "127.0.0.1", port: address.port, agent: false }, incoming => {
                const chunks: Buffer[] = [];
                incoming.on("data", chunk => chunks.push(Buffer.from(chunk)));
                incoming.once("error", reject);
                incoming.once("end", () => {
                    const headers = new Headers();
                    for (const [name, value] of Object.entries(incoming.headers)) {
                        if (value !== undefined) headers.set(name, Array.isArray(value) ? value.join(", ") : value);
                    }
                    resolve(new Response(Buffer.concat(chunks), { status: incoming.statusCode, headers }));
                });
            });
            request.once("error", reject);
        });
    } finally {
        server.closeAllConnections();
        await new Promise<void>(resolve => server.close(() => resolve()));
    }
}
