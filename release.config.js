/**
 * @type {import('semantic-release').GlobalConfig}
 */
module.exports = {
	tagFormat: '${version}',
	branches: ['main', { name: 'trunk', prerelease: true }],
	plugins: [
		[
			'@semantic-release/commit-analyzer',
			{
				preset: 'conventionalcommits',
			},
		],
		'@semantic-release/release-notes-generator',
		// '@semantic-release/changelog',
		[
			'@semantic-release/git',
			{
				message: 'ci(release): release ${nextRelease.version}\n\n${nextRelease.notes}',
			},
		],
		[
			'@semantic-release/gitlab',
			{
				gitlabUrl: 'https://git.stamus-networks.com',
				assets: [
					{
						url: `${process.env.CI_API_V4_URL}/projects/${process.env.CI_PROJECT_ID}/packages/generic/stamusctl/${process.env.NEXT_VERSION}/stamusctl`,
						label: 'stamusctl',
					},
					{
						url: `${process.env.CI_API_V4_URL}/projects/${process.env.CI_PROJECT_ID}/packages/generic/stamusd/${process.env.NEXT_VERSION}/stamusctl`,
						label: 'stamusd',
					},
				],
			},
		],
	],
}
