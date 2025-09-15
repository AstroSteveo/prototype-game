#!/usr/bin/env node

/**
 * Test script for project URL parsing logic
 * This mirrors the parsing logic used in the project-sync.yml workflow
 */

function parseProjectUrl(projectUrl) {
  console.log('🔍 Testing PROJECT_URL:', projectUrl);

  // Extract project info from URL (same regex as workflow)
  const urlMatch = projectUrl.match(
    /github\.com\/(users|orgs)\/([^\/]+)\/projects\/(\d+)/
  );
  
  if (!urlMatch) {
    return {
      success: false,
      error: `❌ Invalid project URL format: ${projectUrl}. Expected format: https://github.com/{users|orgs}/{owner}/projects/{number}`
    };
  }

  const [, scope, owner, projectNumber] = urlMatch;
  const parsedInfo = {
    kind: scope === 'users' ? 'user' : 'org',
    login: owner,
    number: parseInt(projectNumber, 10)
  };
  
  return {
    success: true,
    parsedInfo,
    message: `✅ Parsed PROJECT_URL: kind=${parsedInfo.kind} login=${parsedInfo.login} number=${parsedInfo.number}`
  };
}

// Test cases
const testCases = [
  // Valid URLs - should pass
  { url: 'https://github.com/users/AstroSteveo/projects/2', shouldPass: true, expectedKind: 'user' },
  { url: 'https://github.com/orgs/MyOrg/projects/1', shouldPass: true, expectedKind: 'org' },
  { url: 'https://github.com/users/testuser/projects/123', shouldPass: true, expectedKind: 'user' },
  
  // Invalid URLs - should fail
  { url: 'https://github.com/AstroSteveo/projects/2', shouldPass: false }, // missing users/orgs
  { url: 'https://github.com/users/AstroSteveo/project/2', shouldPass: false }, // project instead of projects
  { url: 'https://github.com/users/AstroSteveo/projects/', shouldPass: false }, // missing number
  { url: 'https://example.com/users/test/projects/1', shouldPass: false }, // wrong domain
  { url: 'invalid-url', shouldPass: false },
  { url: '', shouldPass: false }
];

console.log('🧪 Testing Project URL Parsing Logic\n');

let passCount = 0;
let totalTests = testCases.length;

testCases.forEach((testCase, index) => {
  console.log(`Test ${index + 1}:`);
  
  const result = parseProjectUrl(testCase.url);
  
  if (testCase.shouldPass) {
    // Should be valid
    if (result.success) {
      console.log(result.message);
      if (result.parsedInfo.kind === testCase.expectedKind) {
        console.log(`✅ Correctly identified as ${testCase.expectedKind} project`);
        passCount++;
      } else {
        console.log(`❌ Expected ${testCase.expectedKind} but got ${result.parsedInfo.kind}`);
      }
    } else {
      console.log('❌ Expected valid URL but parsing failed');
      console.log(result.error);
    }
  } else {
    // Should be invalid
    if (!result.success) {
      console.log(result.error);
      console.log('✅ Correctly identified as invalid URL');
      passCount++;
    } else {
      console.log('❌ Expected invalid URL but parsing succeeded');
      console.log(result.message);
    }
  }
  
  console.log('');
});

console.log(`🎯 Test Results: ${passCount}/${totalTests} tests passed`);

if (passCount === totalTests) {
  console.log('🎉 All tests passed!');
  process.exit(0);
} else {
  console.log('❌ Some tests failed');
  process.exit(1);
}