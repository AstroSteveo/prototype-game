#!/usr/bin/env node

/**
 * Demo script showing the improved project sync logic
 * This simulates the workflow behavior with different PROJECT_URL values
 */

function simulateWorkflowLogic(projectUrl, hasToken) {
  console.log('🔄 Simulating workflow execution...');
  console.log('📝 Input conditions:');
  console.log(`   PROJECT_URL: "${projectUrl}"`);
  console.log(`   PROJECTS_TOKEN: ${hasToken ? '✅ provided' : '❌ missing'}`);
  
  // Step 1: Job condition check
  const shouldRun = projectUrl !== '' && hasToken;
  console.log(`\n🚦 Job condition check: ${shouldRun ? '✅ PASS - job will run' : '❌ SKIP - job will be skipped'}`);
  
  if (!shouldRun) {
    console.log('   Reason: Missing PROJECT_URL or PROJECTS_TOKEN');
    return { success: false, reason: 'job_skipped' };
  }
  
  // Step 2: URL parsing
  console.log('\n🔍 URL parsing step...');
  const urlMatch = projectUrl.match(
    /github\.com\/(users|orgs)\/([^\/]+)\/projects\/(\d+)/
  );
  
  if (!urlMatch) {
    const errorMsg = `❌ Invalid project URL format: ${projectUrl}. Expected format: https://github.com/{users|orgs}/{owner}/projects/{number}`;
    console.log(errorMsg);
    return { success: false, reason: 'invalid_url', error: errorMsg };
  }
  
  const [, scope, owner, projectNumber] = urlMatch;
  const parsedInfo = {
    kind: scope === 'users' ? 'user' : 'org',
    login: owner,
    number: parseInt(projectNumber, 10)
  };
  
  console.log(`✅ Parsed PROJECT_URL: kind=${parsedInfo.kind} login=${parsedInfo.login} number=${parsedInfo.number}`);
  
  // Step 3: GraphQL query strategy
  console.log('\n🎯 GraphQL query strategy...');
  const primaryIsUser = parsedInfo.kind === 'user';
  console.log(`   Primary scope: ${parsedInfo.kind}`);
  console.log(`   Fallback scope: ${primaryIsUser ? 'org' : 'user'}`);
  console.log(`   Query order: ${parsedInfo.kind} first, then ${primaryIsUser ? 'org' : 'user'} if needed`);
  
  return { 
    success: true, 
    parsedInfo,
    queryStrategy: {
      primary: parsedInfo.kind,
      fallback: primaryIsUser ? 'org' : 'user'
    }
  };
}

console.log('🎬 Project Sync Workflow Logic Demo\n');

// Test scenarios
const scenarios = [
  {
    name: 'Valid user project (current config)',
    projectUrl: 'https://github.com/users/AstroSteveo/projects/2',
    hasToken: true
  },
  {
    name: 'Valid org project',
    projectUrl: 'https://github.com/orgs/MyCompany/projects/1',
    hasToken: true
  },
  {
    name: 'Missing token',
    projectUrl: 'https://github.com/users/AstroSteveo/projects/2',
    hasToken: false
  },
  {
    name: 'Empty PROJECT_URL',
    projectUrl: '',
    hasToken: true
  },
  {
    name: 'Malformed URL (missing scope)',
    projectUrl: 'https://github.com/AstroSteveo/projects/2',
    hasToken: true
  },
  {
    name: 'Malformed URL (wrong path)',
    projectUrl: 'https://github.com/users/AstroSteveo/project/2',
    hasToken: true
  }
];

scenarios.forEach((scenario, index) => {
  console.log(`\n${'='.repeat(80)}`);
  console.log(`📋 Scenario ${index + 1}: ${scenario.name}`);
  console.log('='.repeat(80));
  
  const result = simulateWorkflowLogic(scenario.projectUrl, scenario.hasToken);
  
  if (result.success) {
    console.log('\n🎉 Workflow would execute successfully!');
    console.log('📊 Expected behavior:');
    console.log(`   1. Query ${result.queryStrategy.primary} project first`);
    console.log(`   2. If not found, fallback to ${result.queryStrategy.fallback} project`);
    console.log('   3. Proceed with project field updates');
  } else {
    console.log(`\n❌ Workflow would fail/skip: ${result.reason}`);
    if (result.error) {
      console.log('   Error details: ' + result.error);
    }
  }
});

console.log(`\n${'='.repeat(80)}`);
console.log('📚 Summary');
console.log('='.repeat(80));
console.log('✅ Improvements made:');
console.log('   • Robust URL parsing with clear error messages');
console.log('   • Smart scope detection (user vs org)');
console.log('   • Fallback query strategy');
console.log('   • Early validation to prevent unnecessary runs');
console.log('   • Detailed debug logging for troubleshooting');
console.log('\n🔧 This should resolve the "Could not resolve to an Organization" error!');