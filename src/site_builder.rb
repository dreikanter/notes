require "fileutils"
require "tilt"
require "uri"
require "yaml"

require_relative "./site"

class SiteBuilder
  attr_reader :configuration

  def initialize(configuration)
    @configuration = configuration
  end

  def build
    site = Site.new(configuration)
    commons = {site: site, configuration: configuration}

    site.pages.each do |page|
      related_pages = site.related_to(page)
      render("page.html", layout: "layout.html", path: page.page_file, locals: commons.merge(page: page, related_pages: related_pages))
      render("redirect.html", path: page.redirect_file, locals: commons.merge(page: page))
    end

    # tags.each do |tag|
    #   render("tag.html", layout: "layout.html", path: "tags/#{tag}/index.html", locals: {pages: pages, tags: tags, current_tag: tag, title: "##{tag}", configuration: configuration, site: site})
    # end

    # render("index.html", layout: "layout.html", path: "index.html", locals: {pages: pages, tags: tags, configuration: configuration, site: site})

    # render("feed.xml", path: "feed.xml", locals: {pages: pages, tags: tags, configuration: configuration, site: site})
  end

  private

  def render(template_name, path:, locals:, layout: nil)
    puts "rendering #{path}"
    File.join(configuration.build_path, path).tap do |path|
      FileUtils.mkdir_p(File.dirname(path))
      File.write(path, render_with_optional_layout(template_name, layout, locals))
    end
  end

  def render_with_optional_layout(template_name, layout_name, locals)
    html = render_template(template_name, locals)
    layout_name ? render_template(layout_name, locals) { html } : html
  end

  def render_template(template_name, locals = {}, &block)
    template(template_name).render(nil, locals.transform_keys(&:to_s), &block)
  end

  def template(template_name)
    @templates ||= {}
    @templates[template_name] ||= Tilt::ERBTemplate.new(template_path(template_name))
  end

  def template_path(template_name)
    File.join(configuration.templates_path, "#{template_name}.erb")
  end
end
