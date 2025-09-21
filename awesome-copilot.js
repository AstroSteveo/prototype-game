#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

const { applyConfig } = require("./apply-config");
const {
  DEFAULT_CONFIG_PATH,
  CONFIG_SECTIONS,
  SECTION_METADATA,
  loadConfig,
  saveConfig,
  ensureConfigStructure,
  countEnabledItems,
  getAllAvailableItems,
  computeEffectiveItemStates,
  extractItemName,
  getSectionFromPath
} = require("./config-manager");

const CONFIG_FLAG_ALIASES = ["--config", "-c"];
const CONTEXT_WARNING_CHAR_LIMIT = {
  instructions: 90000,
  prompts: 45000,
  chatmodes: 30000
};

const numberFormatter = new Intl.NumberFormat("en-US");

const commands = {
  init: {
    description: "Initialize a new project with awesome-copilot configuration",
    usage: "awesome-copilot init [config-file]",
    action: async (args) => {
      const configFile = args[0] || DEFAULT_CONFIG_PATH;
      const { initializeProject } = require("./initialize-project");
      await initializeProject(configFile);
    }
  },

  apply: {
    description: "Apply configuration and copy files to project",
    usage: "awesome-copilot apply [config-file]",
    action: async (args) => {
      const configFile = args[0] || DEFAULT_CONFIG_PATH;
      await applyConfig(configFile);
    }
  },

  list: {
    description: "List items in the configuration with their enabled status",
    usage: "awesome-copilot list [section] [--config <file>]",
    action: (args) => {
      handleListCommand(args);
    }
  },

  toggle: {
    description: "Enable or disable prompts, instructions, chat modes, or collections",
    usage: "awesome-copilot toggle <section> <name|all> [on|off] [--config <file>]",
    action: (args) => {
      handleToggleCommand(args);
    }
  },

  help: {
    description: "Show help information",
    usage: "awesome-copilot help",
    action: () => {
      showHelp();
    }
  }
};

function showHelp() {
  console.log("🤖 Awesome GitHub Copilot Configuration Tool");
  console.log("=".repeat(50));
  console.log("");
  console.log("Usage: awesome-copilot <command> [options]");
  console.log("");
  console.log("Commands:");

  for (const [name, cmd] of Object.entries(commands)) {
    console.log(`  ${name.padEnd(10)} ${cmd.description}`);
    console.log(`  ${" ".repeat(10)} ${cmd.usage}`);
    console.log("");
  }

  console.log("Examples:");
  console.log("  awesome-copilot init                          # Create default config file");
  console.log("  awesome-copilot init my-config.yml            # Create named config file");
  console.log("  awesome-copilot apply                         # Apply default config");
  console.log("  awesome-copilot list instructions             # See which instructions are enabled");
  console.log("  awesome-copilot toggle prompts create-readme on  # Enable a specific prompt");
  console.log("  awesome-copilot toggle instructions all off --config team.yml  # Disable all instructions");
  console.log("");
  console.log("Workflow:");
  console.log("  1. Run 'awesome-copilot init' to create a configuration file");
  console.log("  2. Use 'awesome-copilot list' and 'awesome-copilot toggle' to manage enabled items");
  console.log("  3. Run 'awesome-copilot apply' to copy files to your project");
}

function showError(message) {
  console.error(`❌ Error: ${message}`);
  console.log("");
  console.log("Run 'awesome-copilot help' for usage information.");
  process.exit(1);
}

function handleListCommand(rawArgs) {
  const { args, configPath } = extractConfigOption(rawArgs);

  let sectionsToShow = CONFIG_SECTIONS;
  if (args.length > 0) {
    const requestedSection = validateSectionType(args[0]);
    sectionsToShow = [requestedSection];
  }

  const { config } = loadConfig(configPath);
  const sanitizedConfig = ensureConfigStructure(config);
  const effectiveStates = computeEffectiveItemStates(sanitizedConfig);

  console.log(`📄 Configuration: ${configPath}`);

  sectionsToShow.forEach(section => {
    const availableItems = getAllAvailableItems(section);
    let enabledCount, totalCharacters;
    
    if (section === 'collections') {
      // For collections, use the traditional count
      enabledCount = countEnabledItems(sanitizedConfig[section]);
      totalCharacters = 0;
    } else {
      // For other sections, use effective state count
      enabledCount = effectiveStates[section].size;
      ({ totalCharacters } = calculateSectionFootprint(section, sanitizedConfig[section]));
    }
    
    const headingParts = [
      `${SECTION_METADATA[section].label} (${enabledCount}/${availableItems.length} enabled)`
    ];

    if (totalCharacters > 0 && section !== "collections") {
      headingParts.push(`~${formatNumber(totalCharacters)} chars`);
    }

    console.log(`\n${headingParts.join(", ")}`);

    if (!availableItems.length) {
      console.log("  (no items available)");
      return;
    }

    availableItems.forEach(itemName => {
      let isEnabled, reasonText;
      
      if (section === 'collections') {
        // Collections use the traditional boolean check
        isEnabled = Boolean(sanitizedConfig[section]?.[itemName]);
        reasonText = '';
      } else {
        // Other sections use effective state
        isEnabled = effectiveStates[section].has(itemName);
        const reason = effectiveStates.reasons[section][itemName];
        
        if (reason.source === 'explicit') {
          reasonText = ` (explicit:${reason.value})`;
        } else if (reason.source === 'collections' && reason.via.length > 0) {
          reasonText = ` (via: ${reason.via.join(', ')})`;
        } else {
          reasonText = '';
        }
      }
      
      console.log(`  [${isEnabled ? "✓" : " "}] ${itemName}${reasonText}`);
    });
  });

  console.log("\nUse 'awesome-copilot toggle' to enable or disable specific items.");
}

function handleToggleCommand(rawArgs) {
  const { args, configPath } = extractConfigOption(rawArgs);

  if (args.length < 2) {
    throw new Error("Usage: awesome-copilot toggle <section> <name|all> [on|off] [--config <file>]");
  }

  const section = validateSectionType(args[0]);
  const itemName = args[1];
  const stateArg = args[2];
  const desiredState = stateArg ? parseStateToken(stateArg) : null;

  const availableItems = getAllAvailableItems(section);
  const availableSet = new Set(availableItems);
  if (!availableItems.length) {
    throw new Error(`No ${SECTION_METADATA[section].label.toLowerCase()} available to toggle.`);
  }

  const { config, header } = loadConfig(configPath);
  
  // Compute effective states BEFORE making changes (for all sections that need delta summary)
  const beforeEffective = computeEffectiveItemStates(config);
  
  const configCopy = {
    ...config,
    [section]: { ...config[section] }
  };
  const sectionState = configCopy[section];

  let newState; // Track the new state for delta summary

  if (itemName === "all") {
    if (desiredState === null) {
      throw new Error("Specify 'on' or 'off' when toggling all items.");
    }
    availableItems.forEach(item => {
      sectionState[item] = desiredState;
    });
    console.log(`${desiredState ? "Enabled" : "Disabled"} all ${SECTION_METADATA[section].label.toLowerCase()}.`);

    if (section === "instructions" && desiredState) {
      console.log("⚠️  Enabling every instruction can exceed Copilot Agent's context window. Consider enabling only what you need.");
    }
    newState = desiredState;
  } else {
    if (!availableSet.has(itemName)) {
      const suggestion = findClosestMatch(itemName, availableItems);
      if (suggestion) {
        throw new Error(`Unknown ${SECTION_METADATA[section].singular} '${itemName}'. Did you mean '${suggestion}'?`);
      }
      throw new Error(`Unknown ${SECTION_METADATA[section].singular} '${itemName}'.`);
    }

    const currentState = Boolean(sectionState[itemName]);
    newState = desiredState === null ? !currentState : desiredState;
    sectionState[itemName] = newState;
    console.log(`${newState ? "Enabled" : "Disabled"} ${SECTION_METADATA[section].singular} '${itemName}'.`);
  }

  const sanitizedConfig = ensureConfigStructure(configCopy);
  saveConfig(configPath, sanitizedConfig, header);

  // Show traditional count for all sections
  const enabledCount = countEnabledItems(sanitizedConfig[section]);
  const totalAvailable = availableItems.length;
  const { totalCharacters } = calculateSectionFootprint(section, sanitizedConfig[section]);

  console.log(`${SECTION_METADATA[section].label}: ${enabledCount}/${totalAvailable} enabled.`);
  if (totalCharacters > 0 && section !== "collections") {
    console.log(`Estimated ${SECTION_METADATA[section].label.toLowerCase()} context size: ${formatNumber(totalCharacters)} characters.`);
  }
  maybeWarnAboutContext(section, totalCharacters);

  // For collections, show delta summary of what items are affected
  if (section === 'collections' && itemName !== 'all') {
    const afterEffective = computeEffectiveItemStates(sanitizedConfig);
    showCollectionDelta(beforeEffective, afterEffective, itemName, newState);
  }

  console.log("Run 'awesome-copilot apply' to copy updated selections into your project.");
}

/**
 * Show delta summary for collection toggles
 */
function showCollectionDelta(beforeEffective, afterEffective, collectionName, collectionEnabled) {
  if (!beforeEffective || !afterEffective) return;

  const sections = ['prompts', 'instructions', 'chatmodes'];
  let totalNewlyEnabled = 0;
  let totalNewlyDisabled = 0;
  let totalBlocked = 0;

  sections.forEach(section => {
    const beforeSet = beforeEffective[section];
    const afterSet = afterEffective[section];
    
    // Find newly enabled items
    const newlyEnabled = [];
    afterSet.forEach(item => {
      if (!beforeSet.has(item)) {
        const reason = afterEffective.reasons[section][item];
        if (reason.source === 'collections' && reason.via.includes(collectionName)) {
          newlyEnabled.push(item);
        }
      }
    });

    // Find newly disabled items
    const newlyDisabled = [];
    beforeSet.forEach(item => {
      if (!afterSet.has(item)) {
        const beforeReason = beforeEffective.reasons[section][item];
        if (beforeReason.source === 'collections' && beforeReason.via.includes(collectionName)) {
          newlyDisabled.push(item);
        }
      }
    });

    // Find items blocked by explicit overrides
    const blocked = [];
    if (collectionEnabled) {
      // When enabling a collection, find items that should be enabled but are blocked by explicit:false
      const { parseCollectionYaml } = require("./yaml-parser");
      const collectionPath = path.join(__dirname, "collections", `${collectionName}.collection.yml`);
      if (fs.existsSync(collectionPath)) {
        try {
          const collection = parseCollectionYaml(collectionPath);
          if (collection && collection.items) {
            collection.items.forEach(item => {
              const itemName = extractItemName(item.path);
              const sectionFromPath = getSectionFromPath(item.path);
              if (sectionFromPath === section) {
                const reason = afterEffective.reasons[section][itemName];
                if (reason && reason.source === 'explicit' && reason.value === false) {
                  blocked.push(itemName);
                }
              }
            });
          }
        } catch (error) {
          // Ignore parsing errors
        }
      }
    }

    totalNewlyEnabled += newlyEnabled.length;
    totalNewlyDisabled += newlyDisabled.length;
    totalBlocked += blocked.length;
  });

  // Show concise summary
  if (collectionEnabled) {
    if (totalNewlyEnabled > 0) {
      console.log(`📈 +${totalNewlyEnabled} items effectively enabled by this collection`);
    }
    if (totalBlocked > 0) {
      console.log(`🚫 ${totalBlocked} items remain disabled due to explicit overrides`);
    }
  } else {
    if (totalNewlyDisabled > 0) {
      console.log(`📉 -${totalNewlyDisabled} items effectively disabled by disabling this collection`);
    }
  }
}

function extractConfigOption(rawArgs) {
  const args = [...rawArgs];
  let configPath = DEFAULT_CONFIG_PATH;

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (CONFIG_FLAG_ALIASES.includes(arg)) {
      if (i === args.length - 1) {
        throw new Error("Missing configuration file after --config flag.");
      }
      configPath = args[i + 1];
      args.splice(i, 2);
      i -= 1;
    }
  }

  if (args.length > 0) {
    const potentialPath = args[args.length - 1];
    if (isConfigFilePath(potentialPath)) {
      configPath = potentialPath;
      args.pop();
    }
  }

  return { args, configPath };
}

function isConfigFilePath(value) {
  if (typeof value !== "string") {
    return false;
  }
  return value.endsWith(".yml") || value.endsWith(".yaml") || value.includes("/") || value.includes("\\");
}

function validateSectionType(input) {
  const normalized = String(input || "").toLowerCase();
  if (!SECTION_METADATA[normalized]) {
    throw new Error(`Unknown section '${input}'. Expected one of: ${CONFIG_SECTIONS.join(", ")}.`);
  }
  return normalized;
}

function parseStateToken(token) {
  const normalized = token.toLowerCase();
  if (["on", "enable", "enabled", "true", "yes", "y"].includes(normalized)) {
    return true;
  }
  if (["off", "disable", "disabled", "false", "no", "n"].includes(normalized)) {
    return false;
  }
  throw new Error("State must be 'on' or 'off'.");
}

function calculateSectionFootprint(section, state = {}) {
  const meta = SECTION_METADATA[section];
  if (!meta || section === "collections") {
    return { totalCharacters: 0 };
  }

  let totalCharacters = 0;

  for (const [name, enabled] of Object.entries(state)) {
    if (!enabled) continue;

    const filePath = path.join(__dirname, meta.dir, `${name}${meta.ext}`);
    try {
      const stats = fs.statSync(filePath);
      totalCharacters += stats.size;
    } catch (error) {
      // If the file no longer exists we skip it but continue gracefully.
    }
  }

  return { totalCharacters };
}

function maybeWarnAboutContext(section, totalCharacters) {
  const limit = CONTEXT_WARNING_CHAR_LIMIT[section];
  if (!limit || totalCharacters <= 0) {
    return;
  }

  if (totalCharacters >= limit) {
    console.log(`⚠️  Warning: Estimated ${SECTION_METADATA[section].label.toLowerCase()} size ${formatNumber(totalCharacters)} characters exceeds the recommended limit of ${formatNumber(limit)} characters. Copilot Agent may truncate or crash.`);
  } else if (totalCharacters >= limit * 0.8) {
    console.log(`⚠️  Heads up: Estimated ${SECTION_METADATA[section].label.toLowerCase()} size ${formatNumber(totalCharacters)} characters is approaching the recommended limit (${formatNumber(limit)} characters).`);
  }
}

function formatNumber(value) {
  return numberFormatter.format(Math.round(value));
}

function findClosestMatch(target, candidates) {
  const normalizedTarget = target.toLowerCase();
  return candidates.find(candidate => candidate.toLowerCase().includes(normalizedTarget));
}

async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    showHelp();
    return;
  }

  const command = args[0];
  const commandArgs = args.slice(1);

  if (!commands[command]) {
    showError(`Unknown command: ${command}`);
  }

  try {
    await commands[command].action(commandArgs);
  } catch (error) {
    showError(error.message);
  }
}

if (require.main === module) {
  main();
}

module.exports = { main };
