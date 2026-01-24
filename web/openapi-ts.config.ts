import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
    input: "../api/schema.yaml",
    output: "./src/lib/client",
});
