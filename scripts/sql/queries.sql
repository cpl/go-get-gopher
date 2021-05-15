-- total sources
SELECT COUNT(repo_name) FROM sources;

-- top repo by image count
SELECT repo_name, COUNT(repo_name) AS total
FROM sources
GROUP BY repo_name
ORDER BY total DESC;

-- top domain by image count
SELECT domain, COUNT(domain) AS total
FROM sources
GROUP BY domain
ORDER BY total DESC;


SELECT canon_domain, COUNT(canon_domain) AS total
FROM sources
WHERE canon_domain <> '' AND canon_domain NOT LIKE 'badges%' AND
        canon_domain NOT IN (
                             'img.shields.io',
                             'flat.badgen.net',
                             'opencollective.com',
                             'ci.appveyor.com',
                             'goreportcard.com',
                             'godoc.org',
                             'travis-ci.org',
                             'images.microbadger.com',
                             'codecov.io',
                             'circleci.com',
                             'travis-ci.com',
                             'pkg.go.dev',
                             'badges.gitter.im',
                             'coveralls.io',
                             'app.fossa.io',
                             'asciinema.org'
        ) AND domain NOT LIKE 'avatars%'
GROUP BY canon_domain
ORDER BY total DESC;

-- filtered
SELECT url FROM sources
WHERE canon_domain NOT LIKE 'badges%' AND
        canon_domain NOT IN (
                             'img.shields.io',
                             'flat.badgen.net',
                             'opencollective.com',
                             'ci.appveyor.com',
                             'goreportcard.com',
                             'godoc.org',
                             'travis-ci.org',
                             'images.microbadger.com',
                             'codecov.io',
                             'circleci.com',
                             'travis-ci.com',
                             'pkg.go.dev',
                             'badges.gitter.im',
                             'coveralls.io',
                             'app.fossa.io',
                             'asciinema.org'
        ) AND domain NOT LIKE 'avatars%';

SELECT url FROM sources WHERE alt LIKE '%logo%' OR alt LIKE '%gopher%';

SELECT canon_url FROM sources WHERE canon_domain = 'i.imgur.com';
