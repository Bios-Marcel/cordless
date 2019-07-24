# Copyright 2019 Gentoo Authors
# Distributed under the terms of the GNU General Public License v2

EAPI=7

DESCRIPTION="The Discord terminal client you never knew you wanted."
HOMEPAGE="https://github.com/Bios-Marcel/cordless"

LICENSE="BSD 3-Clause License"
SLOT="0"
KEYWORDS="*"
IUSE=""

DEPEND="=dev-lang/go-*"
RDEPEND="${DEPEND}"
BDEPEND=""

GITHUB_REPO="cordless"
GITHUB_USER="Bios-Maricel"
GITHUB_TAG="f728fc4d741f9219f6523a47c7f42ae23eec4813"
SRC_URI="https://github.com/${GITHUB_USER}/$GITHUB_REPO}/tarball/${GITHUB_TAG} -> ${PN}-${GITHUB_TAG}.tar.gz"

src_unpack() {
	unpack ${A}
	mv "${WORKDIR}/${GITHUB_USER}-${GITHUB_REPO}"-??????? "${S}" || die
}
