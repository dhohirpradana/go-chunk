import fs from "fs";
import path from "path";
import fetch from "node-fetch";
import FormData from "form-data";

// File and upload settings
const filePath = "./to-upload/2024-10-22-07-39-55.zip";
const chunkSize = 1024 * 1024 * 2; // 10MB
const serverUrl = "http://localhost:8080/upload";

// Function to upload a single chunk
async function uploadChunk(fileID, chunkIndex, totalChunks, chunk) {
  const formData = new FormData();
  formData.append("fileID", fileID);
  formData.append("chunkIndex", chunkIndex);
  formData.append("totalChunks", totalChunks);
  formData.append("fileChunk", chunk, `chunk_${chunkIndex}`);

  try {
    const response = await fetch(serverUrl, {
      method: "POST",
      body: formData,
    });

    if (!response.ok) {
      throw new Error(`Error uploading chunk ${chunkIndex}`);
    }

    console.log(`Chunk ${chunkIndex} uploaded successfully`);
  } catch (error) {
    console.error(error);
  }
}

// Function to split file into chunks and upload each one
async function uploadFile() {
  const baseName = path.basename(filePath, path.extname(filePath));
  const extension = path.extname(filePath);
  const fileID = `${baseName}${extension}`;
  const fileStats = fs.statSync(filePath);
  const totalChunks = Math.ceil(fileStats.size / chunkSize);

  console.log(`Starting upload of file '${filePath}' in ${totalChunks} chunks`);

  const fileStream = fs.createReadStream(filePath, {
    highWaterMark: chunkSize,
  });
  let chunkIndex = 0;

  for await (const chunk of fileStream) {
    await uploadChunk(fileID, chunkIndex, totalChunks, chunk);
    chunkIndex++;
  }

  console.log("Upload completed!");
}

uploadFile().catch(console.error);
