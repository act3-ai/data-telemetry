-- SQLite
SELECT id,
    digest
FROM bottles
WHERE bottles.digest = "sha256:05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9";
SELECT manifests.id,
    manifests.bottle_id,
    manifests.digest,
    manifests.manifest
FROM manifests
    INNER JOIN bottles ON bottles.id = manifests.bottle_id
WHERE bottles.digest = "sha256:05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9";
SELECT events.repository
FROM events
    INNER JOIN manifests ON manifests.id = events.manifest_id
WHERE manifests.bottle_id = 2;
SELECT events.repository
FROM events
    INNER JOIN manifests ON manifests.id = events.manifest_id
    INNER JOIN bottles ON bottles.id = manifests.bottle_id
WHERE bottles.digest = "sha256:05a8efd3483c60a4364d3f6f328ee1897facdbffb043b51941424a34121bbbe9";
SELECT manifest_id
FROM events;
SELECT id,
    data,
    LENGTH("data") AS SizeOfData
FROM blobs
WHERE SizeOfData < 31;
-- Find bottles with multiple selectors
SELECT DISTINCT metrics.name,
    metrics.description,
    metrics.value,
    bottles.digest
FROM `bottles`
    INNER JOIN metrics ON metrics.bottle_id = bottles.id
    INNER JOIN labels label0x0 ON label0x0.bottle_id = bottles.id
    INNER JOIN labels label0x1 ON label0x1.bottle_id = bottles.id
WHERE metrics.name = "training loss"
    AND (
        label0x0.key = "mykey"
        AND label0x0.value = "myvalue"
    )
    AND (
        label0x1.key = "myotherkey"
        AND label0x1.value = "myothervalue2"
    )
    AND `bottles`.`deleted_at` IS NULL
UNION
SELECT DISTINCT metrics.name,
    metrics.description,
    metrics.value,
    bottles.digest
FROM `bottles`
    INNER JOIN metrics ON metrics.bottle_id = bottles.id
    INNER JOIN labels label1x0 ON label1x0.bottle_id = bottles.id
WHERE metrics.name = "training loss"
    AND (
        label1x0.key = "mykey"
        AND label1x0.value = "doesnotexist"
    )
    AND `bottles`.`deleted_at` IS NULL
ORDER BY metrics.value ASC
LIMIT 7;
-- Round Two
SELECT DISTINCT metrics.name,
    metrics.description,
    metrics.value,
    bottles.digest
FROM `bottles`
    INNER JOIN metrics ON metrics.bottle_id = bottles.id
    INNER JOIN labels label0x0 ON label0x0.bottle_id = bottles.id
    INNER JOIN labels label0x1 ON label0x1.bottle_id = bottles.id
    INNER JOIN labels label1x0 ON label1x0.bottle_id = bottles.id
WHERE metrics.name = "training loss"
    AND (
        (
            (
                label0x0.key = "mykey"
                AND label0x0.value = "myvalue"
            )
            AND (
                label0x1.key = "myotherkey"
                AND label0x1.value = "myothervalue2"
            )
        )
        OR (
            label1x0.key = "mykey"
            AND label1x0.value = "doesnotexist"
        )
    )
    AND `bottles`.`deleted_at` IS NULL
ORDER BY metrics.value ASC
LIMIT 7;
-- handleBottle manifestations query
SELECT m.repository,
    m.auth_required,
    m.digest,
    m.last_accessed_at
FROM (
        SELECT events.repository,
            events.auth_required,
            events.manifest_digest AS digest,
            events.timestamp AS last_accessed_at,
            RANK() OVER (
                PARTITION BY events.repository,
                events.manifest_digest
                ORDER BY events.timestamp DESC
            ) rank
        FROM `events`
            INNER JOIN bottles ON events.bottle_id = bottles.id
            INNER JOIN manifests ON events.bottle_id = manifests.id
            INNER JOIN digests ON digests.data_id = bottles.data_id
            AND digests.digest = "sha256:594f360554cd8e114add71dbaa17a604adcc262ab876c8a0afc56d0ff030b496"
    ) as m
WHERE m.rank = 1
ORDER BY m.last_accessed_at DESC;

-- Find ancestors
SELECT data_id FROM digests WHERE digest="sha256:3e8e2e3db7a8e23283cd0a8bd14d697dcc55459da33c57af5148ec7923924a52";
SELECT id FROM bottles WHERE data_id=9; -- (data_id=9 from above);
SELECT bottle_digest FROM sources WHERE bottle_id = 2; -- (id=2 from above);
SELECT data_id FROM digests WHERE digest IN ("sha256:93648e4272e7d8044959f7d64d75175425099f4fd9eb80f7e0ea294e0034fdef");---(bottle_digest from above);
SELECT * FROM bottles WHERE data_id IN (8); -- (data_id=(8) from above);

SELECT DISTINCT sources.bottle_digest
    FROM bottles
        INNER JOIN digests digests_original ON bottles_original.data_id = digests_original.data_id
        INNER JOIN bottles bottles_original ON digests_original.data_id = bottles_original.data_id
        INNER JOIN sources ON bottles_original.id = sources.bottle_id
    WHERE digests_original.digest = "sha256:3e8e2e3db7a8e23283cd0a8bd14d697dcc55459da33c57af5148ec7923924a52"
;

SELECT bottles.id, bottles.description
FROM bottles
    INNER JOIN digests digests_ancestor ON bottles.data_id = digests_ancestor.data_id
WHERE digests_ancestor.digest IN (
    SELECT DISTINCT sources.bottle_digest
    FROM bottles
        INNER JOIN digests digests_original ON bottles_original.data_id = digests_original.data_id
        INNER JOIN bottles bottles_original ON digests_original.data_id = bottles_original.data_id
        INNER JOIN sources ON bottles_original.id = sources.bottle_id
    WHERE digests_original.digest = "sha256:3e8e2e3db7a8e23283cd0a8bd14d697dcc55459da33c57af5148ec7923924a52"
)
;

-- Find descendents
SELECT data_id FROM digests WHERE digest="sha256:93648e4272e7d8044959f7d64d75175425099f4fd9eb80f7e0ea294e0034fdef";
SELECT digest AS aliases FROM digests where data_id=8; -- (data_id from above)
SELECT bottle_id FROM sources WHERE bottle_digest IN ("sha256:93648e4272e7d8044959f7d64d75175425099f4fd9eb80f7e0ea294e0034fdef"); --(aliases from above, all known aliases of the original bottle)
SELECT * FROM bottles WHERE id = (2); --(bottle_id from above);

SELECT DISTINCT digests.digest
    FROM digests
        INNER JOIN digests digests_original ON digests.data_id = digests_original.data_id
    WHERE digests_original.digest = "sha256:93648e4272e7d8044959f7d64d75175425099f4fd9eb80f7e0ea294e0034fdef"
;

SELECT bottles.id, bottles.description
FROM bottles
    INNER JOIN sources ON sources.bottle_id = bottles.id
WHERE bottle_digest IN (
    SELECT DISTINCT digests.digest
    FROM digests
        INNER JOIN digests digests_original ON digests.data_id = digests_original.data_id
    WHERE digests_original.digest = "sha256:93648e4272e7d8044959f7d64d75175425099f4fd9eb80f7e0ea294e0034fdef"
);
