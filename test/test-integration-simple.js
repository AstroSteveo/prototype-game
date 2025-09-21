#!/usr/bin/env node

/**
 * Simple integration test for toggle + apply workflow
 */

const assert = require('assert');
const path = require('path');
const fs = require('fs');
const { execSync } = require('child_process');

console.log('🧪 Integration test: Basic toggle + apply workflow...\n');

function test(name, fn) {
  try {
    fn();
    console.log(`✅ ${name}`);
  } catch (error) {
    console.log(`❌ ${name}: ${error.message}`);
    process.exit(1);
  }
}

// Test 1: List command shows effective states and reasons
test('List command shows effective states and reasons', () => {
  const output = execSync('node awesome-copilot.js list prompts', { 
    encoding: 'utf8',
    cwd: path.join(__dirname, '..')
  });
  
  assert(output.includes('(explicit:'), 'Should show explicit reasons');
  assert(output.includes('(via:') || output.includes('✓'), 'Should show collection reasons or enabled items');
});

// Test 2: Toggle command shows delta summaries for collections
test('Toggle collection shows delta summary', () => {
  const output = execSync('node awesome-copilot.js toggle collections azure-cloud-development off', { 
    encoding: 'utf8',
    cwd: path.join(__dirname, '..')
  });
  
  // Re-enable it
  const output2 = execSync('node awesome-copilot.js toggle collections azure-cloud-development on', { 
    encoding: 'utf8',
    cwd: path.join(__dirname, '..')
  });
  
  assert(output2.includes('Collection') || output2.includes('📈') || output2.includes('🚫'), 'Should show some kind of delta summary');
});

// Test 3: Apply command uses effective states
test('Apply command works with effective states', () => {
  const output = execSync('node apply-config.js', { 
    encoding: 'utf8',
    cwd: path.join(__dirname, '..')
  });
  
  assert(output.includes('Configuration applied successfully!'), 'Apply should complete successfully');
  assert(output.includes('📝 Total files copied:'), 'Should show copy summary');
});

// Test 4: Effective state computation handles precedence correctly
test('Effective state computation handles precedence', () => {
  const { computeEffectiveItemStates } = require(path.join(__dirname, '..', 'config-manager'));
  
  const config = {
    collections: { 'testing-automation': true },
    prompts: { 'playwright-generate-test': false }, // explicit false overrides collection
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  
  // Should be disabled due to explicit false, despite collection being enabled
  assert(!effective.prompts.has('playwright-generate-test'), 'Explicit false should override collection');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'explicit');
});

console.log('\n🎉 All integration tests passed! The workflow is working correctly.');