// ============================================
// TaskFlow - Prisma Configuration
// Uses DIRECT_URL for migrations (port 5432)
// Uses DATABASE_URL for runtime queries (port 6543, pooled)
// ============================================
import "dotenv/config";
import { defineConfig } from "prisma/config";

export default defineConfig({
  schema: "prisma/schema.prisma",
  migrations: {
    path: "prisma/migrations",
  },
  datasource: {
    url: process.env["DIRECT_URL"],
  },
});
