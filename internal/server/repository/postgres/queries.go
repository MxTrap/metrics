package postgres

const updateStmt = `UPDATE metric SET 
	metric_type_id = (SELECT id FROM metric_type WHERE metric_type = $1), 
    metric_name = $2, 
    value=$3,
    delta = delta + $4
    WHERE metric_name = $2;`

const insertStmt = `INSERT INTO metric (metric_type_id, metric_name, value, delta)
						VALUES ((SELECT id FROM metric_type WHERE metric_type = $1), $2, $3, $4);`

const findStmt = `SELECT m.id, t.metric_type, m.metric_name, m.value, m.delta FROM metric AS m
JOIN metric_type AS t ON m.metric_type_id = t.id WHERE m.metric_name = $1;`

const selectAllStmt = `SELECT m.id, t.metric_type, m.metric_name, m.value, m.delta FROM metric AS m 
    		JOIN metric_type AS t ON m.metric_type_id = t.id;`
