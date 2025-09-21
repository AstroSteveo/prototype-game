#!/usr/bin/env node

/**
 * Basic unit tests for computeEffectiveItemStates function
 * Tests the core logic for collection precedence and effective state computation
 */

const assert = require('assert');
const path = require('path');
const fs = require('fs');

// Add parent directory to path to require config-manager
const parentDir = path.join(__dirname, '..');
const { computeEffectiveItemStates } = require(path.join(parentDir, 'config-manager'));

console.log('🧪 Testing computeEffectiveItemStates function...\n');

function test(name, fn) {
  try {
    fn();
    console.log(`✅ ${name}`);
  } catch (error) {
    console.log(`❌ ${name}: ${error.message}`);
    process.exit(1);
  }
}

// Test 1: Explicit true overrides collections off
test('Explicit true overrides collections being disabled', () => {
  const config = {
    collections: {
      'testing-automation': false
    },
    prompts: {
      'playwright-generate-test': true
    },
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  assert(effective.prompts.has('playwright-generate-test'), 'Item should be enabled by explicit true');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'explicit');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].value, true);
});

// Test 2: Explicit false overrides collections on
test('Explicit false overrides collections being enabled', () => {
  const config = {
    collections: {
      'testing-automation': true
    },
    prompts: {
      'playwright-generate-test': false
    },
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  assert(!effective.prompts.has('playwright-generate-test'), 'Item should be disabled by explicit false');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'explicit');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].value, false);
});

// Test 3: Undefined inherits from enabled collection
test('Undefined item inherits from enabled collection', () => {
  const config = {
    collections: {
      'testing-automation': true
    },
    prompts: {
      // playwright-generate-test: undefined (not specified)
    },
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  assert(effective.prompts.has('playwright-generate-test'), 'Item should be enabled via collection');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'collections');
  assert(effective.reasons.prompts['playwright-generate-test'].via.includes('testing-automation'));
});

// Test 4: Undefined with no enabled collections = disabled
test('Undefined item with no enabled collections is disabled', () => {
  const config = {
    collections: {
      'testing-automation': false
    },
    prompts: {
      // playwright-generate-test: undefined (not specified)
    },
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  assert(!effective.prompts.has('playwright-generate-test'), 'Item should be disabled with no collections');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'default');
});

// Test 5: Item in multiple collections - enabled if ANY collection is enabled
test('Item enabled if ANY collection containing it is enabled', () => {
  const config = {
    collections: {
      'frontend-web-dev': true,
      'testing-automation': false
    },
    prompts: {
      // playwright-generate-test: undefined (both collections contain this item)
    },
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  assert(effective.prompts.has('playwright-generate-test'), 'Item should be enabled via frontend-web-dev');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'collections');
  assert(effective.reasons.prompts['playwright-generate-test'].via.includes('frontend-web-dev'));
});

// Test 6: Multiple enabled collections show multiple in 'via' array
test('Multiple enabled collections show in via array', () => {
  const config = {
    collections: {
      'frontend-web-dev': true,
      'testing-automation': true
    },
    prompts: {
      // playwright-generate-test: undefined (both collections contain this item)
    },
    instructions: {},
    chatmodes: {}
  };
  
  const effective = computeEffectiveItemStates(config);
  assert(effective.prompts.has('playwright-generate-test'), 'Item should be enabled');
  assert.strictEqual(effective.reasons.prompts['playwright-generate-test'].source, 'collections');
  const via = effective.reasons.prompts['playwright-generate-test'].via;
  assert(via.includes('frontend-web-dev') && via.includes('testing-automation'), 'Should list both collections');
});

// Test 7: Different sections work independently
test('Different sections work independently', () => {
  const config = {
    collections: {
      'testing-automation': true
    },
    prompts: {
      'playwright-generate-test': false // explicit false
    },
    instructions: {
      // playwright-typescript: undefined (inherits from collection)
    },
    chatmodes: {
      // playwright-tester: undefined (inherits from collection)
    }
  };
  
  const effective = computeEffectiveItemStates(config);
  
  // Prompt is explicitly disabled
  assert(!effective.prompts.has('playwright-generate-test'), 'Prompt should be disabled by explicit false');
  
  // Instruction inherits from collection
  assert(effective.instructions.has('playwright-typescript'), 'Instruction should be enabled via collection');
  
  // Chat mode inherits from collection  
  assert(effective.chatmodes.has('playwright-tester'), 'Chat mode should be enabled via collection');
});

console.log('\n🎉 All tests passed! The computeEffectiveItemStates function is working correctly.');