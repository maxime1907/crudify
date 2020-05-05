CREATE TABLE "crudify" (
	"id" integer NOT NULL,
	"name" VARCHAR(255) NOT NULL UNIQUE,
	"creation" DATE NOT NULL,
	"description" TEXT NOT NULL,
	"admin" BOOLEAN NOT NULL DEFAULT 'false',
	CONSTRAINT crudify_pk PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);