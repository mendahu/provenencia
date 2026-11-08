-- S14-40: catalog_search_inputs records the inputs of each person, event and place summary (and each association's ends) so a write can find the rows it moved. Owned by searchindex; no foreign key, so a deleted input keeps its dependents' rows until they are reprojected.
CREATE TABLE catalog_search_inputs (
    dependent_id BLOB NOT NULL,
    input_id BLOB NOT NULL,
    bridge TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (dependent_id, input_id, bridge, endpoint)
) STRICT, WITHOUT ROWID;

CREATE INDEX catalog_search_inputs_input_idx ON catalog_search_inputs (input_id, bridge, endpoint);
