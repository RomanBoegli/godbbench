\benchmark once \name initialize
MATCH (n) DETACH DELETE n

// INSERT
\benchmark loop 0.75 \name insert_customer
CREATE (n:Customer {CustomerId: {{.Iter}}, Name: '{{call .RandString 3 10 }}', Address: '{{call .RandString 10 50 }}', Birthday: '{{call .RandDate }}' });

\benchmark loop 1.0 \name insert_order
CREATE (n:Order {OrderId: {{.Iter}}, CreationDate: '{{call .RandDate }}', Comment: '{{call .RandString 10 50 }}' });
MATCH (a:Customer), (b:Order) WHERE a.CustomerId = {{.Iter}} AND b.OrderId = {{.Iter}} CREATE (a)-[r:PLACES]->(b);

\benchmark loop 0.25 \name insert_supplier
CREATE (n:Supplier {SupplierId: {{.Iter}}, Name: '{{call .RandString 3 10 }}', Address: '{{call .RandString 10 50 }}' });

\benchmark loop 0.1 \name insert_category
CREATE (n:Category {CategoryId: {{.Iter}}, Name: '{{call .RandString 3 10 }}' });

\benchmark loop 0.5 \name insert_product
CREATE (n:Product { ProductId: {{.Iter}}, Code: '{{call .RandString 5 6 }}', Description: '{{call .RandString 0 100 }}', UnitSize: {{call .RandIntBetween 1 10 }}, PricePerUnit: {{call .RandFloatBetween 0.01 999999.99 }} });
MATCH (a:Supplier), (b:Product) WHERE a.SupplierId = {{.Iter}} AND b.ProductId = {{.Iter}} CREATE (a)-[r:SUPPLIES]->(b);
MATCH (a:Category), (b:Product) WHERE a.CategoryId = {{.Iter}} AND b.ProductId = {{.Iter}} CREATE (a)-[r:GROUPS]->(b);

\benchmark loop 1.0 \name insert_lineitem
CREATE (n:LineItem {LineItemId: {{.Iter}}, Quantity: 123, DeliveryDate: '{{call .RandDate }}' });
MATCH (a:Order), (b:LineItem) WHERE a.OrderId = {{.Iter}} AND b.LineItemId = {{.Iter}} CREATE (a)-[r:CONTAINS]->(b);
MATCH (a:Product), (b:LineItem) WHERE a.ProductId = {{.Iter}} AND b.LineItemId = {{.Iter}} CREATE (a)-[r:OCCURS]->(b);


// SELECTS
\benchmark loop 1.0 \name select_simple
MATCH (n:Customer) WHERE n.CustomerId = {{.Iter}} RETURN n

\benchmark loop 1.0 \name select_medium
MATCH (n:Product)<-[r:SUPPLIES]-(f:Supplier) WHERE f.SupplierId = {{.Iter}} RETURN n ORDER BY n.PricePerUnit DESC

\benchmark loop 1.0 \name select_complex
MATCH (c:Customer)-[:PLACES]->(o:Order)-[:CONTAINS]->(l:LineItem)<-[:OCCURS]-(p:Product) WHERE '{{call .RandDate }}' <= o.CreationDate <= "9999-12-31" RETURN c.CustomerId, c.Name, SUM(l.Quantity * p.UnitSize * p.PricePerUnit) as TotalOrderValue ORDER BY TotalOrderValue DESC


// CLEAN
\benchmark once \name clean
MATCH (n) DETACH DELETE n;
Call db.clearQueryCaches();