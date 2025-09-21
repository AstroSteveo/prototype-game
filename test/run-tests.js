#!/usr/bin/env node

/**
 * Test runner for effective-state collections functionality
 */

console.log('🧪 Running all tests for effective-state collections...\n');

try {
  // Run unit tests
  console.log('📋 Running unit tests...');
  require('./test-effective-states.js');
  
  console.log('\n📋 Running integration tests...');
  require('./test-integration-simple.js');
  
  console.log('\n🎉 All tests passed successfully!');
  console.log('\n✅ Core functionality working:');
  console.log('  - Effective state computation with proper precedence');
  console.log('  - CLI shows reasons (explicit vs via collections)');
  console.log('  - Collection toggles show delta summaries');
  console.log('  - Apply uses effective states correctly');
  console.log('  - Shared items protected when any collection still requires them');
  
} catch (error) {
  console.error('\n❌ Test suite failed:', error.message);
  process.exit(1);
}