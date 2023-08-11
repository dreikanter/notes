require_relative "./page"

class Site
  attr_reader :configuration

  def initialize(configuration)
    @configuration = configuration
  end

  def pages
    @pages ||= note_files.map { Page.new(source_file: _1, configuration: configuration) }.filter(&:public?)
  end

  def tags
    @tags ||= pages.map(&:tags).flatten.uniq
  end

  def related_to(page)
    pages.filter { _1.uid != page.uid && _1.tags.intersect?(page.tags) }
  end

  private

  def note_files
    Dir.glob("#{configuration.notes_path}/**/*.md")
  end
end
