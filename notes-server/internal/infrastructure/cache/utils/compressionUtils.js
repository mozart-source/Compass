const { promisify } = require('util');
const { gzip, gunzip } = require('zlib');
const { logger } = require('../../../../pkg/utils/logger');

const gzipAsync = promisify(gzip);
const gunzipAsync = promisify(gunzip);

function toBase64(buffer) {
  return buffer.toString('base64');
}

function fromBase64(str) {
  return Buffer.from(str, 'base64');
}

function safeStringify(obj) {
  try {
    return JSON.stringify(obj);
  } catch (e) {
    logger.error('JSON stringify failed', { error: e.message });
    return '';
  }
}

function safeParse(str) {
  try {
    return JSON.parse(str);
  } catch (e) {
    logger.error('JSON parse failed', { error: e.message });
    return null;
  }
}

async function compress(data) {
  try {
    const buffer = await gzipAsync(safeStringify(data));
    return toBase64(buffer);
  } catch (error) {
    logger.error('Compression failed:', { error: error.message });
    return data;
  }
}

async function decompress(data) {
  try {
    const buffer = fromBase64(data);
    const decompressed = await gunzipAsync(buffer);
    return safeParse(decompressed.toString());
  } catch (error) {
    logger.error('Decompression failed:', { error: error.message });
    return data;
  }
}

module.exports = { compress, decompress, toBase64, fromBase64, safeStringify, safeParse }; 