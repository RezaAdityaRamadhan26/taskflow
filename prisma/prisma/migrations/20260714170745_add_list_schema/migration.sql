-- CreateTable
CREATE TABLE "lists" (
    "id" UUID NOT NULL,
    "board_id" UUID NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "position" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMPTZ NOT NULL,

    CONSTRAINT "lists_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "lists_board_id_idx" ON "lists"("board_id");

-- AddForeignKey
ALTER TABLE "lists" ADD CONSTRAINT "lists_board_id_fkey" FOREIGN KEY ("board_id") REFERENCES "boards"("id") ON DELETE CASCADE ON UPDATE CASCADE;
