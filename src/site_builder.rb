require "fileutils"
require "tilt"
require "uri"
require "yaml"
require_relative "./site"

class SiteBuilder
  def build
    site.pages.each { render(_1) }
  end

  private

  def site
    @site ||= Site.new
  end

  def render(page)
    local_path = page.local_path
    puts "rendering #{local_path}"
    locals = {configuration: Configuration, site: site, page: page}
    content = render_page(page: page, template: page.template, layout: page.layout, locals: locals)
    write(local_path: local_path, content: content)
  end

  def write(local_path:, content:)
    path = File.join(Configuration.build_path, local_path)
    FileUtils.mkdir_p(File.dirname(path))
    File.write(path, content)
  end

  def render_page(page:, template:, layout:, locals:)
    content = render_template(template: template, locals: locals)
    layout ? render_template(template: layout, locals: locals) { content } : content
  end

  def render_template(template:, locals:, &block)
    load_template(template).render(nil, locals.transform_keys(&:to_s), &block)
  end

  def load_template(name)
    @templates ||= {}
    @templates[name] ||= Tilt::ERBTemplate.new(template_path(name))
  end

  def template_path(template_name)
    File.join(Configuration.templates_path, "#{template_name}.erb")
  end
end
