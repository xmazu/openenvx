/**
 * Database introspection using postgres client
 * Direct connection to PostgreSQL for schema discovery
 */

import postgres from 'postgres';

export interface ColumnInfo {
  dataType: string;
  defaultValue: string | null;
  isNullable: boolean;
  isPrimaryKey: boolean;
  name: string;
}

export interface ForeignKeyInfo {
  column: string;
  foreignColumn: string;
  foreignTable: string;
}

export interface TableSchema {
  columns: ColumnInfo[];
  foreignKeys: ForeignKeyInfo[];
  name: string;
  primaryKey: string | null;
}

let sql: postgres.Sql | null = null;

function getClient(): postgres.Sql {
  if (!sql) {
    const databaseUrl = process.env.DATABASE_URL;
    if (!databaseUrl) {
      throw new Error('DATABASE_URL environment variable is required');
    }
    sql = postgres(databaseUrl);
  }
  return sql;
}

/**
 * Fetch all tables from public schema
 */
export async function fetchTables(): Promise<string[]> {
  const client = getClient();
  const tables = await client`
    SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_type = 'BASE TABLE'
    ORDER BY table_name
  `;
  return tables.map((t) => t.table_name);
}

/**
 * Fetch columns for a specific table
 */
export async function fetchColumns(tableName: string): Promise<ColumnInfo[]> {
  const client = getClient();

  const [columns, pkResult] = await Promise.all([
    client`
      SELECT column_name, data_type, is_nullable, column_default
      FROM information_schema.columns
      WHERE table_schema = 'public' AND table_name = ${tableName}
      ORDER BY ordinal_position
    `,
    client`
      SELECT kcu.column_name
      FROM information_schema.table_constraints tc
      JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name
      WHERE tc.constraint_type = 'PRIMARY KEY'
        AND tc.table_schema = 'public'
        AND tc.table_name = ${tableName}
      LIMIT 1
    `,
  ]);

  const primaryKey = pkResult[0]?.column_name || null;

  return columns.map((row) => ({
    name: row.column_name,
    dataType: row.data_type,
    isNullable: row.is_nullable === 'YES',
    defaultValue: row.column_default,
    isPrimaryKey: row.column_name === primaryKey,
  }));
}

/**
 * Fetch foreign keys for a table
 */
export async function fetchForeignKeys(
  tableName: string
): Promise<ForeignKeyInfo[]> {
  const client = getClient();

  const fks = await client`
    SELECT
      kcu.column_name,
      ccu.table_name AS foreign_table_name,
      ccu.column_name AS foreign_column_name
    FROM information_schema.table_constraints AS tc
    JOIN information_schema.key_column_usage AS kcu 
      ON tc.constraint_name = kcu.constraint_name
    JOIN information_schema.constraint_column_usage AS ccu 
      ON ccu.constraint_name = tc.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
      AND tc.table_schema = 'public'
      AND tc.table_name = ${tableName}
  `;

  return fks.map((row) => ({
    column: row.column_name,
    foreignTable: row.foreign_table_name,
    foreignColumn: row.foreign_column_name,
  }));
}

/**
 * Fetch complete schema for a table
 */
export async function fetchTableSchema(
  tableName: string
): Promise<TableSchema> {
  const [columns, foreignKeys] = await Promise.all([
    fetchColumns(tableName),
    fetchForeignKeys(tableName),
  ]);

  const primaryKey = columns.find((c) => c.isPrimaryKey)?.name || null;

  return {
    name: tableName,
    columns,
    foreignKeys,
    primaryKey,
  };
}

/**
 * Fetch all table schemas
 */
export async function fetchAllSchemas(
  excludeTables: string[] = []
): Promise<TableSchema[]> {
  const tables = await fetchTables();
  const filteredTables = tables.filter((t) => !excludeTables.includes(t));

  const schemas = await Promise.all(
    filteredTables.map((table) => fetchTableSchema(table))
  );

  return schemas;
}

export async function fetchReferenceData(
  tableName: string,
  searchTerm?: string,
  limit = 50,
  displayField = 'name'
): Promise<Record<string, unknown>[]> {
  const client = getClient();

  const columns = await fetchColumns(tableName);
  const pkColumn = columns.find((c) => c.isPrimaryKey)?.name || 'id';

  const columnsToFetch = new Set([pkColumn]);

  if (columns.some((c) => c.name === displayField)) {
    columnsToFetch.add(displayField);
  }

  const columnList = Array.from(columnsToFetch)
    .map((c) => `"${c}"`)
    .join(', ');

  if (searchTerm) {
    const searchPattern = `%${searchTerm}%`;
    const searchCondition = `"${displayField}"::text ILIKE '${searchPattern}'`;

    const results = await client.unsafe(`
      SELECT ${columnList}
      FROM "${tableName}"
      WHERE ${searchCondition}
      LIMIT ${limit}
    `);
    return results as unknown as Record<string, unknown>[];
  }

  const results = await client.unsafe(`
    SELECT ${columnList}
    FROM "${tableName}"
    LIMIT ${limit}
  `);
  return results as unknown as Record<string, unknown>[];
}
