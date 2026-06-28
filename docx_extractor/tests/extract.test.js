const request = require("supertest");
const path = require("path");
const fs = require("fs");
const app = require("../index");

const DOCX_MIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document";
const SAMPLE_DOCX = path.join(__dirname, "testdata", "sample.docx");

describe("POST /extract", () => {

    test("extracts text from valid docx", async () => {
        const res = await request(app)
            .post("/extract")
            .attach("file", SAMPLE_DOCX, { contentType: DOCX_MIME })

        expect(res.status).toBe(200);
        expect(res.body.text).toBeDefined();
        expect(typeof res.body.text).toBe("string");
        expect(res.body.text.length).toBeGreaterThan(0);
        expect(res.body.text).toContain("Hello from test DOCX");
        expect(res.body.text).toContain("Second paragraph");
    });

    test("rejects non-docx content type", async () => {
        const res = await request(app)
            .post("/extract")
            .attach("file", SAMPLE_DOCX, { contentType: "application/pdf" });

        expect(res.status).toBe(415);
    });

    test("rejects plain text content type", async () => {
        const res = await request(app)
            .post("/extract")
            .attach("file", Buffer.from("hello"), { filename: "test.txt", contentType: "text/plain" });

        expect(res.status).toBe(415);
    });

    test("returns 422 when no file provided", async () => {
        const res = await request(app).post("/extract");
        expect(res.status).toBe(422);
    });

    test("returns 500 for corrupted docx", async () => {
        const res = await request(app)
            .post("/extract")
            .attach("file", Buffer.from("not a docx"), { filename: "broken.docx", contentType: DOCX_MIME });

        expect(res.status).toBe(500);
    });

    test("returns 413 for file exceeding size limit", async () => {
        const bigBuffer = Buffer.alloc(21 * 1024 * 1024);

        const res = await request(app)
            .post("/extract")
            .attach("file", bigBuffer, { filename: "big.docx", contentType: DOCX_MIME });

        expect(res.status).toBe(413);
    });

});