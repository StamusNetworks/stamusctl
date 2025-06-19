const gitlab = [
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
]

const github = [
	'@semantic-release/github',
	{
		assets: [
			{
				path: `stamusctl-${process.env.OS}-${process.env.ARCH}`,
				label: `stamusctl (${process.env.OS?.toUpperCase()} ${process.env.ARCH?.toUpperCase()})`,
			},
			{
				path: `stamusd-${process.env.OS}-${process.env.ARCH}`,
				label: `stamusd (${process.env.OS?.toUpperCase()} ${process.env.ARCH?.toUpperCase()})`,
			},
		],
	},
]

// Determine the CI environment
const isGitLab = !!process.env.CI && !process.env.GITHUB_ACTIONS // GitLab CI sets CI, GitHub Actions sets GITHUB_ACTIONS

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
		[
			'@semantic-release/git',
			{
				message: 'ci(release): release ${nextRelease.version}\n\n${nextRelease.notes}',
			},
		],
		...(isGitLab
			? []
			: [
					[
						'@semantic-release/exec',
						{
							verifyReleaseCmd:
								'echo "version=${nextRelease.version}" >> $GITHUB_OUTPUT',
						},
					],
				]),
		...(isGitLab ? [gitlab] : [github]), // Dynamically select GitLab or GitHub
	],
}
