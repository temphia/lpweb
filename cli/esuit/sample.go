package main

var TestUploadData = []byte(`Arch Linux is an independently developed, x86-64 general-purpose GNU/Linux.`)

var TestUploadData2 = []byte(`

Arch Linux is an independently developed, x86-64 general-purpose GNU/Linux distribution that strives to provide the latest stable versions of most software by following a rolling-release model. The default installation is a minimal base system, configured by the user to only add what is purposely required.
Principles
Simplicity

Arch Linux defines simplicity as without unnecessary additions or modifications. It ships software as released by the original developers (upstream) with minimal distribution-specific (downstream) changes: patches not accepted by upstream are avoided, and Arch's downstream patches consist almost entirely of backported bug fixes that are obsoleted by the project's next release.

In a similar fashion, Arch ships the configuration files provided by upstream with changes limited to distribution-specific issues like adjusting the system file paths. It does not add automation features such as enabling a service simply because the package was installed. Packages are only split when compelling advantages exist, such as to save disk space in particularly bad cases of waste. GUI configuration utilities are not officially provided, encouraging users to perform most system configuration from the shell and a text editor.
Modernity

Arch Linux strives to maintain the latest stable release versions of its software as long as systemic package breakage can be reasonably avoided. It is based on a rolling-release system, which allows a one-time installation with continuous upgrades.

Arch incorporates many of the newer features available to GNU/Linux users, including the systemd init system, modern file systems, LVM2, software RAID, udev support and initcpio (with mkinitcpio), as well as the latest available kernels.
Pragmatism

Arch is a pragmatic distribution rather than an ideological one. The principles here are only useful guidelines. Ultimately, design decisions are made on a case-by-case basis through developer consensus. Evidence-based technical analysis and debate are what matter, not politics or popular opinion.

The large number of packages and build scripts in the various Arch Linux repositories offer free and open source software for those who prefer it, as well as proprietary software packages for those who embrace functionality over ideology.
User centrality

Whereas many GNU/Linux distributions attempt to be more user-friendly, Arch Linux has always been, and shall always remain user-centric. The distribution is intended to fill the needs of those contributing to it, rather than trying to appeal to as many users as possible. It is targeted at the proficient GNU/Linux user, or anyone with a do-it-yourself attitude who is willing to read the documentation, and solve their own problems.

All users are encouraged to participate and contribute to the distribution. Reporting and helping fix bugs is highly valued and patches improving packages or the core projects are very appreciated: Arch's developers are volunteers and active contributors will often find themselves becoming part of that team. Archers can freely contribute packages to the Arch User Repository, improve the ArchWiki documentation, provide technical assistance to others or just exchange opinions in the forums, mailing lists, or IRC channels. Arch Linux is the operating system of choice for many people around the globe, and there exist several international communities that offer help and provide documentation in many different languages.
Versatility

Arch Linux is a general-purpose distribution. Upon installation, only a command-line environment is provided; rather than tearing out unneeded and unwanted packages, the user is offered the ability to build a custom system by choosing among thousands of high-quality packages provided in the official repositories for the x86-64 architecture.

Arch is a rolling-release model backed by pacman, a lightweight, simple and fast package manager that allows for continuously upgrading the entire system with one command. Arch also provides the Arch build system, a ports-like system to make it easy to build and install packages from source, which can also be synchronized with one command. In addition, the Arch User Repository contains many thousands of community-contributed PKGBUILD scripts for compiling installable packages from source using the makepkg application. It is also possible for users to build and maintain their own custom repositories with ease.
History

The Arch community has grown and matured to become one of the most popular and influential Linux distributions, also testified by the attention and review received over the years.

Arch developers remain unpaid, part-time volunteers, and there are no prospects for monetizing Arch Linux, so it will remain free in all senses of the word. Those curious to peruse more detail about Arch's development history can browse the Arch entry in the Internet Archive Wayback Machine and the Arch Linux News Archives.
The early years

Judd Vinet, a Canadian programmer and occasional guitarist, began developing Arch Linux in early 2001. Its first formal release, Arch Linux 0.1, was on March 11, 2002. Inspired by the elegant simplicity of Slackware, BSD, PLD Linux and CRUX, and yet disappointed with their lack of package management at the time, Vinet built his own distribution on similar principles as those distros. But, he also wrote a package management program called pacman, to automatically handle package dependency resolution, installation, removal, and upgrades.
The middle years

The early Arch community grew steadily, as evidenced by this chart of forum posts, users, and bug reports. Moreover, it was from its early days known as an open, friendly, and helpful community.
Birth of the ArchWiki

On 2005-07-08 the ArchWiki was first set up on the MediaWiki engine.
The dawning of the age of A. Griffin

In late 2007, Judd Vinet retired from active participation as an Arch developer, and smoothly transferred the reins over to American programmer Aaron Griffin, also known as Phrakture.
Arch Install Scripts

The 2012-07-15 release of the installation image deprecated the menu-driven Arch Installation Framework (AIF) in favor of the Arch Install Scripts (arch-install-scripts).
The systemd era

Between 2012 and 2013 the traditional System V init system was replaced by systemd.[1][2][3][4]
Drop of i686 support

On 2017-01-25 it was announced that support for the i686 architecture would be phased out due to its decreasing popularity among the developers and the community. By the end of November 2017, all i686 packages were removed from the mirrors.
Review of Project Leader role and election

At the start of 2020, in a team effort the Arch Linux staff devised a new process for determining future leaders, documented in DeveloperWiki:Project Leader.

As Aaron Griffin had decided to step down from his role, a poll was held to elect a new person to replace him, and on 2020-02-24 its results were published, making the election of Levente Polyak official.
The GitLab era

In May 2023, Arch Linux migrated its packaging infrastructure to GitLab. Besides internal changes and innovations, this also resulted in splitting the testing repository into core-testing and extra-testing, the staging repository into core-staging and extra-staging, and finally community has been merged into extra. Read more details on the GitLab blog.

Several months later, in November 2023, the old bug tracker (Flyspray) has been migrated to GitLab and its collaboration features (issues and merge requests) have been open for public. For archiving reasons there will be a static copy of the old bug tracker so that links (for example the randomly picked FS#56716) are still valid. 

`)
