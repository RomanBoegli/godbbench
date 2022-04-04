\benchmark once \name initialize
MATCH (n) DETACH DELETE n

// INSERT
\benchmark loop 0.75 \name insert_customer
CREATE (n:Customer {CustomerId: {{.Iter}}, Name: '{{call .RandString 3 10 }}', Address: '{{call .RandString 10 50 }}', Birthday: '{{call .RandDate }}' });

\benchmark loop 1.0 \name insert_order
CREATE (n:Order {OrderId: {{.Iter}}, CreationDate: '{{call .RandDate }}', Comment: '{{call .RandString 10 50 }}' });
CALL {MATCH (x:Customer) RETURN x, rand() as rand ORDER BY rand ASC LIMIT 1} WITH x.CustomerId as fk MATCH (a:Customer), (b:Order) WHERE a.CustomerId = fk AND b.OrderId = {{.Iter}} CREATE (a)-[r:PLACES]->(b);

\benchmark loop 0.25 \name insert_supplier
CREATE (n:Supplier {SupplierId: {{.Iter}}, Name: '{{call .RandString 3 10 }}', Address: '{{call .RandString 10 50 }}' });

\benchmark loop 0.1 \name insert_category
CREATE (n:Category {CategoryId: {{.Iter}}, Name: '{{call .RandString 3 10 }}' });

\benchmark loop 0.5 \name insert_product
CREATE (n:Product { ProductId: {{.Iter}}, Code: '{{call .RandString 5 6 }}', Description: '{{call .RandString 0 100 }}', UnitSize: {{call .RandIntBetween 1 10 }}, PricePerUnit: {{call .RandFloatBetween 0.01 999999.99 }} });
CALL {MATCH (x:Supplier) RETURN x, rand() as rand ORDER BY rand ASC LIMIT 1} WITH x.SupplierId as fk MATCH (a:Supplier), (b:Product) WHERE a.SupplierId = fk AND b.ProductId = {{.Iter}} CREATE (a)-[r:SUPPLIES]->(b);
CALL {MATCH (x:Category) RETURN x, rand() as rand ORDER BY rand ASC LIMIT 1} WITH x.CategoryId as fk MATCH (a:Category), (b:Product) WHERE a.CategoryId = fk AND b.ProductId = {{.Iter}} CREATE (a)-[r:GROUPS]->(b);

\benchmark loop 1.0 \name insert_lineitem
CREATE (n:LineItem {LineItemId: {{.Iter}}, Quantity: 123, DeliveryDate: '{{call .RandDate }}' });
CALL {MATCH (x:Order) RETURN x, rand() as rand ORDER BY rand ASC LIMIT 1} WITH x.OrderId as fk MATCH (a:Order), (b:LineItem) WHERE a.OrderId = fk AND b.LineItemId = {{.Iter}} CREATE (a)-[r:CONTAINS]->(b);
CALL {MATCH (x:Product) RETURN x, rand() as rand ORDER BY rand ASC LIMIT 1} WITH x.ProductId as fk MATCH (a:Product), (b:LineItem) WHERE a.ProductId = fk AND b.LineItemId = {{.Iter}} CREATE (a)-[r:OCCURS]->(b);


// SELECTS

\benchmark loop 1.0 \name select_simple
MATCH (n:Customer) WHERE n.CustomerId = {{.Iter}} RETURN n

\benchmark loop 1.0 \name select_medium
MATCH (n:Product)<-[r:SUPPLIES]-(f:Supplier) WHERE f.SupplierId = {{.Iter}} RETURN n ORDER BY n.PricePerUnit DESC

\benchmark loop 1.0 \name select_complex
MATCH (c:Customer)-[:PLACES]->(o:Order)-[:CONTAINS]->(l:LineItem)<-[:OCCURS]-(p:Product) WHERE '{{call .RandDate }}' <= o.CreationDate <= "9999-12-31" RETURN c.CustomerId, c.Name, SUM(l.Quantity * p.UnitSize * p.PricePerUnit) as TotalOrderValue ORDER BY TotalOrderValue DESC


// DROPS
//\benchmark once \name clean
//MATCH (n) DETACH DELETE n
