/**
 * xID → EpixNet Relayer
 *
 * Queries on-chain xID state digest and attestations, fetches domain data,
 * then publishes it to the EpixNet site's chain data directory.
 *
 * Usage:
 *   npx ts-node relayer/index.ts --rpc http://localhost:8545 --site-dir /path/to/epixnet/site
 */

import { createPublicClient, http, type Abi, parseAbi } from "viem";
import * as fs from "fs";
import * as path from "path";
import * as crypto from "crypto";

// xID precompile address
const XID_PRECOMPILE = "0x0000000000000000000000000000000000000900" as const;

// ABI for the attestation-related precompile methods
const xidAbi = parseAbi([
  "function getStateDigest() view returns (string digest, uint64 height, uint64 numNames)",
  "function getAttestations() view returns (string[] validators, string[] signatures, uint64[] heights, bool finalized)",
  "function resolve(string name, string tld) view returns (address owner)",
  "function getProfile(string name, string tld) view returns (string avatar, string bio)",
  "function getEpixNetPeers(string name, string tld) view returns (string[] addresses, string[] labels, uint64[] addedAts, bool[] actives, uint64[] revokedAts)",
  "function getContentRoot(string name, string tld) view returns (string root, uint64 updatedAt)",
]);

interface Config {
  rpcUrl: string;
  siteDir: string;
  chainDataDir: string;
  pollIntervalMs: number;
}

function parseArgs(): Config {
  const args = process.argv.slice(2);
  let rpcUrl = "http://localhost:8545";
  let siteDir = "";
  let chainDataDir = "data/chain/epixchain";
  let pollIntervalMs = 30000;

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case "--rpc":
        rpcUrl = args[++i];
        break;
      case "--site-dir":
        siteDir = args[++i];
        break;
      case "--chain-data-dir":
        chainDataDir = args[++i];
        break;
      case "--poll-interval":
        pollIntervalMs = parseInt(args[++i]) * 1000;
        break;
    }
  }

  if (!siteDir) {
    console.error("Usage: npx ts-node relayer/index.ts --rpc <url> --site-dir <path>");
    process.exit(1);
  }

  return { rpcUrl, siteDir, chainDataDir, pollIntervalMs };
}

function sha512(data: string | Buffer): string {
  return crypto.createHash("sha512").update(data).digest("hex");
}

async function main() {
  const config = parseArgs();

  const client = createPublicClient({
    transport: http(config.rpcUrl),
  });

  console.log(`xID Relayer starting...`);
  console.log(`  RPC: ${config.rpcUrl}`);
  console.log(`  Site: ${config.siteDir}`);
  console.log(`  Chain data dir: ${config.chainDataDir}`);

  let lastDigest = "";

  while (true) {
    try {
      // 1. Query current state digest
      const [digest, height, numNames] = await client.readContract({
        address: XID_PRECOMPILE,
        abi: xidAbi,
        functionName: "getStateDigest",
      });

      if (!digest || digest === lastDigest) {
        await sleep(config.pollIntervalMs);
        continue;
      }

      console.log(`\nNew digest: ${digest} (height=${height}, names=${numNames})`);

      // 2. Query attestations
      const [validators, signatures, heights, finalized] = await client.readContract({
        address: XID_PRECOMPILE,
        abi: xidAbi,
        functionName: "getAttestations",
      });

      if (!finalized) {
        console.log(`  Digest not yet finalized (${validators.length} attestations). Waiting...`);
        await sleep(config.pollIntervalMs);
        continue;
      }

      console.log(`  Finalized with ${validators.length} attestations`);

      // 3. Fetch all domain data
      // Note: This is a simplified approach - for production, use paginated queries
      // For now we query individual domains that we know about
      const domains = await fetchAllDomains(client, numNames);

      // 4. Write domain files
      const outputDir = path.join(config.siteDir, config.chainDataDir);
      const domainsDir = path.join(outputDir, "domains");
      fs.mkdirSync(domainsDir, { recursive: true });

      const files: Record<string, { sha512: string; size: number }> = {};

      for (const domain of domains) {
        const filename = `domains/${domain.name}.${domain.tld}.json`;
        const filePath = path.join(outputDir, filename);
        const content = JSON.stringify(domain, null, 2);
        fs.writeFileSync(filePath, content);
        files[filename] = {
          sha512: sha512(content),
          size: Buffer.byteLength(content),
        };
      }

      // 5. Write meta.json
      const meta = {
        chain_id: "epixchain-1",
        block_height: Number(height),
        digest: digest,
        num_names: Number(numNames),
        timestamp: Math.floor(Date.now() / 1000),
      };
      const metaContent = JSON.stringify(meta, null, 2);
      fs.writeFileSync(path.join(outputDir, "meta.json"), metaContent);
      files["meta.json"] = {
        sha512: sha512(metaContent),
        size: Buffer.byteLength(metaContent),
      };

      // 6. Build content.json with validator signatures over the state_digest
      const signs: Record<string, string> = {};
      for (let i = 0; i < validators.length; i++) {
        signs[validators[i]] = signatures[i];
      }

      const contentJson = {
        inner_path: `${config.chainDataDir}/content.json`,
        modified: Math.floor(Date.now() / 1000),
        state_digest: digest,
        files,
        signs,
      };

      const contentJsonStr = JSON.stringify(contentJson, null, 2);
      fs.writeFileSync(path.join(outputDir, "content.json"), contentJsonStr);

      console.log(`  Published ${domains.length} domains to ${outputDir}`);
      lastDigest = digest;

    } catch (err) {
      console.error("Relayer error:", err);
    }

    await sleep(config.pollIntervalMs);
  }
}

interface DomainData {
  name: string;
  tld: string;
  owner: string;
  profile?: { avatar: string; bio: string };
  peers: Array<{ address: string; label: string; addedAt: number; active: boolean; revokedAt: number }>;
  contentRoot?: string;
}

async function fetchAllDomains(
  client: any,
  numNames: bigint
): Promise<DomainData[]> {
  // In a production relayer, you'd use the gRPC QueryStateSnapshot endpoint
  // or paginate through all names. For now, this is a placeholder that
  // indicates where the domain fetching logic would go.
  console.log(`  TODO: Fetch ${numNames} domains via QueryStateSnapshot gRPC`);
  console.log(`  For production, use the chain's gRPC endpoint for paginated domain data`);
  return [];
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

main().catch(console.error);
