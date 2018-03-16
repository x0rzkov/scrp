-- To run:
-- cqlsh --ssl -f schema.1.cql 

-- SCRP
drop keyspace scrp;
create keyspace scrp WITH REPLICATION = {  'class':'NetworkTopologyStrategy', 'DC1':'1' };

use scrp;


create table slugs (
  id uuid,
  seq timeuuid,
  key text,
  PRIMARY KEY ((id),seq)
);
--CREATE INDEX adfsdf ON ddf(fdf)