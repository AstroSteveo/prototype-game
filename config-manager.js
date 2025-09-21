const fs = require("fs");
const path = require("path");

const { parseConfigYamlContent } = require("./apply-config");
const { objectToYaml, generateConfigHeader, getAvailableItems } = require("./generate-config");

const DEFAULT_CONFIG_PATH = "awesome-copilot.config.yml";
const SECTION_METADATA = {
  prompts: { dir: "prompts", ext: ".prompt.md", label: "Prompts", singular: "prompt" },
  instructions: { dir: "instructions", ext: ".instructions.md", label: "Instructions", singular: "instruction" },
  chatmodes: { dir: "chatmodes", ext: ".chatmode.md", label: "Chat Modes", singular: "chat mode" },
  collections: { dir: "collections", ext: ".collection.yml", label: "Collections", singular: "collection" }
};
const CONFIG_SECTIONS = Object.keys(SECTION_METADATA);

function loadConfig(configPath = DEFAULT_CONFIG_PATH) {
  if (!fs.existsSync(configPath)) {
    throw new Error(`Configuration file not found: ${configPath}`);
  }

  const rawContent = fs.readFileSync(configPath, "utf8");
  const { header, body } = splitHeaderAndBody(rawContent);
  const parsed = parseConfigYamlContent(body || "");
  const config = ensureConfigStructure(parsed || {});

  return { config, header };
}

function saveConfig(configPath, config, header) {
  const ensuredConfig = ensureConfigStructure(config || {});
  const sortedConfig = sortConfigSections(ensuredConfig);
  const yamlContent = objectToYaml(sortedConfig);
  const headerContent = formatHeader(header);

  fs.writeFileSync(configPath, headerContent + yamlContent);
}

function splitHeaderAndBody(content) {
  const lines = content.split("\n");
  const headerLines = [];
  let firstBodyIndex = 0;

  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (trimmed === "" || trimmed.startsWith("#")) {
      headerLines.push(lines[i]);
      firstBodyIndex = i + 1;
    } else {
      firstBodyIndex = i;
      break;
    }
  }

  const header = headerLines.join("\n");
  const body = lines.slice(firstBodyIndex).join("\n");

  return { header, body };
}

function ensureConfigStructure(config) {
  const sanitized = typeof config === "object" && config !== null ? { ...config } : {};

  if (!sanitized.version) {
    sanitized.version = "1.0";
  }

  const project = typeof sanitized.project === "object" && sanitized.project !== null ? { ...sanitized.project } : {};
  if (project.output_directory === undefined) {
    project.output_directory = ".awesome-copilot";
  }
  sanitized.project = project;

  CONFIG_SECTIONS.forEach(section => {
    sanitized[section] = sanitizeSection(sanitized[section]);
  });

  return sanitized;
}

function sanitizeSection(section) {
  if (!section || typeof section !== "object") {
    return {};
  }

  const sanitized = {};
  for (const [key, value] of Object.entries(section)) {
    sanitized[key] = toBoolean(value);
  }

  return sanitized;
}

function toBoolean(value) {
  if (typeof value === "boolean") {
    return value;
  }

  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    if (normalized === "true") return true;
    if (normalized === "false") return false;
  }

  return Boolean(value);
}

function sortConfigSections(config) {
  const sorted = { ...config };

  CONFIG_SECTIONS.forEach(section => {
    sorted[section] = sortObjectKeys(sorted[section]);
  });

  return sorted;
}

function sortObjectKeys(obj) {
  if (!obj || typeof obj !== "object") {
    return {};
  }

  return Object.keys(obj)
    .sort((a, b) => a.localeCompare(b))
    .reduce((acc, key) => {
      acc[key] = obj[key];
      return acc;
    }, {});
}

function formatHeader(existingHeader) {
  const header = existingHeader && existingHeader.trim().length > 0
    ? existingHeader
    : generateConfigHeader();

  let normalized = header;

  if (!normalized.endsWith("\n")) {
    normalized += "\n";
  }
  if (!normalized.endsWith("\n\n")) {
    normalized += "\n";
  }

  return normalized;
}

function countEnabledItems(section = {}) {
  return Object.values(section).filter(Boolean).length;
}

function getAllAvailableItems(type) {
  const meta = SECTION_METADATA[type];

  if (!meta) {
    return [];
  }

  return getAvailableItems(path.join(__dirname, meta.dir), meta.ext);
}

/**
 * Compute effective item states based on explicit flags and enabled collections
 * @param {Object} config - The full configuration object
 * @returns {Object} - Object with effectively enabled Sets per section and reason metadata
 */
function computeEffectiveItemStates(config) {
  const result = {
    prompts: new Set(),
    instructions: new Set(),
    chatmodes: new Set(),
    reasons: {
      prompts: {},
      instructions: {},
      chatmodes: {}
    }
  };

  // First, gather items from enabled collections
  const { parseCollectionYaml } = require("./yaml-parser");
  const collectionMembership = {
    prompts: new Map(),
    instructions: new Map(),
    chatmodes: new Map()
  };

  if (config.collections) {
    for (const [collectionName, enabled] of Object.entries(config.collections)) {
      if (enabled) {
        const collectionPath = path.join(__dirname, "collections", `${collectionName}.collection.yml`);
        if (fs.existsSync(collectionPath)) {
          try {
            const collection = parseCollectionYaml(collectionPath);
            if (collection && collection.items) {
              collection.items.forEach(item => {
                const itemName = extractItemName(item.path);
                const section = getSectionFromPath(item.path);
                
                if (section && collectionMembership[section]) {
                  if (!collectionMembership[section].has(itemName)) {
                    collectionMembership[section].set(itemName, new Set());
                  }
                  collectionMembership[section].get(itemName).add(collectionName);
                }
              });
            }
          } catch (error) {
            console.warn(`Warning: Failed to parse collection ${collectionName}: ${error.message}`);
          }
        }
      }
    }
  }

  // Now compute effective states for each section
  const sections = ['prompts', 'instructions', 'chatmodes'];
  sections.forEach(section => {
    const sectionConfig = config[section] || {};
    const availableItems = getAllAvailableItems(section);

    availableItems.forEach(itemName => {
      const explicitFlag = sectionConfig[itemName];
      const inCollections = collectionMembership[section].get(itemName);

      // Apply precedence rules:
      // 1. Explicit boolean overrides collections
      // 2. If no explicit flag (undefined), inherit from collections
      // 3. Enabled if in any enabled collection, disabled otherwise
      
      let isEffectivelyEnabled = false;
      let reason = { source: 'default', via: [] };

      if (explicitFlag === true) {
        isEffectivelyEnabled = true;
        reason = { source: 'explicit', value: true };
      } else if (explicitFlag === false) {
        isEffectivelyEnabled = false;
        reason = { source: 'explicit', value: false };
      } else {
        // undefined - inherit from collections
        if (inCollections && inCollections.size > 0) {
          isEffectivelyEnabled = true;
          reason = { source: 'collections', via: Array.from(inCollections).sort() };
        } else {
          isEffectivelyEnabled = false;
          reason = { source: 'default' };
        }
      }

      if (isEffectivelyEnabled) {
        result[section].add(itemName);
      }
      result.reasons[section][itemName] = reason;
    });
  });

  return result;
}

/**
 * Extract item name from path (removes directory and extension)
 */
function extractItemName(itemPath) {
  const basename = path.basename(itemPath);
  // Remove the extension (e.g., .prompt.md, .instructions.md, .chatmode.md)
  return basename.replace(/\.(prompt|instructions|chatmode)\.md$/, '');
}

/**
 * Determine section from item path
 */
function getSectionFromPath(itemPath) {
  if (itemPath.includes('prompts/') && itemPath.endsWith('.prompt.md')) {
    return 'prompts';
  } else if (itemPath.includes('instructions/') && itemPath.endsWith('.instructions.md')) {
    return 'instructions';
  } else if (itemPath.includes('chatmodes/') && itemPath.endsWith('.chatmode.md')) {
    return 'chatmodes';
  }
  return null;
}

module.exports = {
  DEFAULT_CONFIG_PATH,
  CONFIG_SECTIONS,
  SECTION_METADATA,
  loadConfig,
  saveConfig,
  splitHeaderAndBody,
  ensureConfigStructure,
  sortObjectKeys,
  countEnabledItems,
  getAllAvailableItems,
  computeEffectiveItemStates
};
