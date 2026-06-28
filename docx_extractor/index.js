const express = require("express");
const mammoth = require("mammoth");
const multer  = require("multer");

const app = express();
const upload = multer({
    storage: multer.memoryStorage(),
    limits: { fileSize: 20 * 1024 * 1024 },
});

const DOCX_MIME = "application/vnd.openxmlformats-officedocument.wordprocessingml.document";

app.post("/extract", upload.single("file"), async (req, res) => {
    if (!req.file) {
        return res.status(422).json({ detail: "file is required" });
    }

    if (req.file.mimetype !== DOCX_MIME) {
        return res.status(415).json({ detail: "this service only handles DOCX files" });
    }

    try {
        const result = await mammoth.extractRawText({ buffer: req.file.buffer });
        return res.json({ text: result.value.trim() });
    } catch (err) {
        return res.status(500).json({ detail: "failed to extract DOCX" });
    }
});

app.use((err, req, res, next) => {
    if (err.code === "LIMIT_FILE_SIZE") {
        return res.status(413).json({ detail: "DOCX exceeds maximum size" });
    }
    return res.status(500).json({ detail: "internal server error" });
});

if (require.main === module) {
  const PORT = process.env.PORT || 3000;
  app.listen(PORT);
}

module.exports = app;