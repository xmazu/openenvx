import type {
  NestedResourceConfig,
  ResourceItem,
  ResourcesConfig,
} from '@/types/resources';

function buildNestedRoutes(
  parentName: string,
  nested: Record<string, NestedResourceConfig> | undefined
): ResourceItem['nested'] {
  if (!nested) {
    return undefined;
  }

  const result: NonNullable<ResourceItem['nested']> = {};

  for (const [name, config] of Object.entries(nested)) {
    result[name] = {
      name,
      label: config.label,
      icon: config.icon,
      list: `/${parentName}/:id/${name}`,
      create: `/${parentName}/:id/${name}/create`,
      edit: `/${parentName}/:id/${name}/:nestedId/edit`,
      show: `/${parentName}/:id/${name}/:nestedId`,
      parentField: config.parentField,
      meta: {
        label: config.label,
        icon: config.icon,
      },
    };
  }

  return result;
}

export function defineResources(config: ResourcesConfig): ResourceItem[] {
  const resources: ResourceItem[] = [];

  for (const [name, resourceConfig] of Object.entries(config)) {
    const resource: ResourceItem = {
      name,
      label: resourceConfig.label,
      icon: resourceConfig.icon,
      list: `/${name}`,
      create: `/${name}/create`,
      edit: `/${name}/:id/edit`,
      show: `/${name}/:id`,
      meta: {
        label: resourceConfig.label,
        icon: resourceConfig.icon,
        description: resourceConfig.description,
      },
      nested: buildNestedRoutes(name, resourceConfig.nested),
      displayField: resourceConfig.displayField ?? 'name',
      config: resourceConfig,
    };

    resources.push(resource);
  }

  return resources;
}
